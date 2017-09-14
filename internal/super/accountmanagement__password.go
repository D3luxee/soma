package super // import "github.com/mjolnir42/soma/internal/super"

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mjolnir42/scrypth64"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/internal/stmt"
	"github.com/mjolnir42/soma/lib/auth"
	uuid "github.com/satori/go.uuid"
)

func (s *Supervisor) userPassword(q *msg.Request) {
	result := msg.FromRequest(q)
	result.Super.Verdict = 401

	var (
		cred                                                  *svCredential
		err                                                   error
		kex                                                   *auth.Kex
		plain                                                 []byte
		timer                                                 *time.Timer
		token                                                 auth.Token
		tx                                                    *sql.Tx
		validFrom, expiresAt, credExpiresAt, credDeactivateAt time.Time
		userID                                                string
		userUUID                                              uuid.UUID
		mcf                                                   scrypth64.Mcf
		ok, active                                            bool
	)
	data := q.Super.Encrypted.Data

	if s.readonly {
		result.ReadOnly()
		goto returnImmediate
	}

	timer = time.NewTimer(1 * time.Second)
	defer timer.Stop()

	// get kex
	if kex = s.kex.read(q.Super.Encrypted.KexID); kex == nil {
		result.Forbidden(fmt.Errorf(`Key exchange not found`))
		goto dispatch
	}

	if !kex.IsSameSourceExtractedString(q.RemoteAddr) {
		result.Forbidden(fmt.Errorf(`Key exchange not found`))
		goto dispatch
	}

	s.kex.remove(q.Super.Encrypted.KexID)

	if err = kex.DecodeAndDecrypt(&data, &plain); err != nil {
		result.ServerError(err)
		goto dispatch
	}

	if err = json.Unmarshal(plain, &token); err != nil {
		result.ServerError(err)
		goto dispatch
	}
	// token.UserName is the username
	// token.Password is the _NEW_ password that should be set
	// token.Token    is either:
	// -- the old  password (change)
	// -- the ldap password (reset/ldap)
	// -- the token         (reset/mailtoken)

	s.reqLog.Printf(msg.LogStrSRq, q.Section, q.Action, token.UserName, q.RemoteAddr)

	if err = s.stmtFindUserID.QueryRow(token.UserName).
		Scan(&userID); err == sql.ErrNoRows {
		result.Forbidden(fmt.Errorf("Unknown user: %s", token.UserName))
		goto dispatch
	} else if err != nil {
		result.ServerError(err)
	}
	userUUID, _ = uuid.FromString(userID)

	// user has to be active
	if err = s.stmtCheckUserActive.QueryRow(userID).
		Scan(&active); err == sql.ErrNoRows {
		result.Forbidden(fmt.Errorf("Unknown user: %s", token.UserName))
		goto dispatch
	}
	if !active {
		result.Forbidden(fmt.Errorf("User %s (%s) is not active", token.UserName, userID))
		goto dispatch
	}

	// change of password or reset of password?
	switch q.Action {
	case `reset`:
		switch s.activation {
		case `ldap`:
			if ok, err = validateLdapCredentials(token.UserName, token.Token); err != nil {
				result.ServerError(err)
				goto dispatch
			} else if !ok {
				result.Forbidden(fmt.Errorf(`Invalid LDAP credentials`))
				goto dispatch
			}
		case `token`:
			result.NotImplemented(fmt.Errorf(`Mail-Token not supported yet`))
			goto dispatch
		default:
			result.ServerError(fmt.Errorf("Unknown activation: %s",
				s.conf.Auth.Activation))
			goto dispatch
		}
	case `change`:
		if cred = s.credentials.read(token.UserName); cred == nil {
			result.Forbidden(fmt.Errorf("Unknown user: %s", token.UserName))
			goto dispatch
		}
		if !cred.isActive {
			result.Forbidden(fmt.Errorf("Inactive user: %s", token.UserName))
			goto dispatch
		}
		if time.Now().UTC().Before(cred.validFrom.UTC()) ||
			time.Now().UTC().After(cred.expiresAt.UTC()) {
			result.Forbidden(fmt.Errorf("Expired: %s", token.UserName))
			goto dispatch
		}
		if ok, err = scrypth64.Verify(token.Token, cred.cryptMCF); err != nil {
			result.ServerError(err)
			goto dispatch
		} else if !ok {
			result.Forbidden(fmt.Errorf(`Invalid credentials`))
			goto dispatch
		}
	default:
		result.NotImplemented(fmt.Errorf("Unknown requested action %s", q.Action))
		goto dispatch
	}
	// OK: validation success

	if mcf, err = scrypth64.Digest(token.Password, nil); err != nil {
		result.Forbidden(err)
		goto dispatch
	}

	token.SetIPAddressExtractedString(q.RemoteAddr)
	if err = token.Generate(mcf, s.key, s.seed); err != nil {
		result.ServerError(err)
		goto dispatch
	}
	validFrom, _ = time.Parse(msg.RFC3339Milli, token.ValidFrom)
	expiresAt, _ = time.Parse(msg.RFC3339Milli, token.ExpiresAt)
	credDeactivateAt = validFrom.Add(time.Second * -1).UTC()
	credExpiresAt = validFrom.Add(time.Duration(s.credExpiry) * time.Hour * 24).UTC()

	// Open transaction to update credentials
	if tx, err = s.conn.Begin(); err != nil {
		result.ServerError(err)
		goto dispatch
	}
	defer tx.Rollback()

	// Invalidate existing credentials
	if _, err = tx.Exec(
		stmt.InvalidateUserCredential,
		credDeactivateAt,
		userUUID,
	); err != nil {
		result.ServerError(err)
		goto dispatch
	}

	// Insert new credentials
	if _, err = tx.Exec(
		stmt.SetUserCredential,
		userUUID,
		mcf.String(),
		validFrom.UTC(),
		credExpiresAt.UTC(),
	); err != nil {
		result.ServerError(err)
		goto dispatch
	}

	// Insert issued token
	if _, err = tx.Exec(
		stmt.InsertToken,
		token.Token,
		token.Salt,
		validFrom.UTC(),
		expiresAt.UTC(),
	); err != nil {
		result.ServerError(err)
		goto dispatch
	}

	// Update supervisor credentialscache
	s.credentials.revoke(token.UserName)
	s.credentials.insert(token.UserName,
		userUUID,
		validFrom.UTC(),
		credExpiresAt.UTC(),
		mcf,
	)
	if err = s.tokens.insert(token.Token,
		token.ValidFrom,
		token.ExpiresAt,
		token.Salt,
	); err != nil {
		result.ServerError(err)
		goto dispatch
	}

	// commit transaction
	if err = tx.Commit(); err != nil {
		result.ServerError(err)
		goto dispatch
	}
	// -> sdata = kex.EncryptAndEncode(&token)
	plain = []byte{}
	data = []byte{}
	if plain, err = json.Marshal(token); err != nil {
		result.ServerError(err)
		goto dispatch
	}
	if err = kex.EncryptAndEncode(&plain, &data); err != nil {
		result.ServerError(err)
		goto dispatch
	}
	// -> send sdata reply
	result.Super.Verdict = 200
	result.Super.Encrypted.Data = data
	result.OK()

dispatch:
	<-timer.C

returnImmediate:
	q.Reply <- result
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
