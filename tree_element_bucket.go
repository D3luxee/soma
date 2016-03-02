package somatree

import (
	"fmt"

	"github.com/satori/go.uuid"
)

type SomaTreeElemBucket struct {
	Id              uuid.UUID
	Name            string
	Environment     string
	Type            string
	State           string
	Frozen          bool
	Deleted         bool
	Repository      uuid.UUID
	Team            uuid.UUID
	Parent          SomaTreeBucketReceiver `json:"-"`
	Fault           *SomaTreeElemFault     `json:"-"`
	PropertyOncall  map[string]SomaTreeProperty
	PropertyService map[string]SomaTreeProperty
	PropertySystem  map[string]SomaTreeProperty
	PropertyCustom  map[string]SomaTreeProperty
	Checks          map[string]Check
	Children        map[string]SomaTreeBucketAttacher //`json:"-"`
	Action          chan *Action                      `json:"-"`
}

type BucketSpec struct {
	Id          string
	Name        string
	Environment string
	Team        string
	Repository  string
	Deleted     bool
	Frozen      bool
}

//
// NEW
func NewBucket(spec BucketSpec) *SomaTreeElemBucket {
	if !specBucketCheck(spec) {
		fmt.Printf("%#v\n", spec) // XXX DEBUG
		panic(`No.`)
	}

	teb := new(SomaTreeElemBucket)
	teb.Id, _ = uuid.FromString(spec.Id)
	teb.Name = spec.Name
	teb.Team, _ = uuid.FromString(spec.Team)
	teb.Environment = spec.Environment
	teb.Frozen = spec.Frozen
	teb.Deleted = spec.Deleted
	teb.Repository, _ = uuid.FromString(spec.Repository)
	teb.Type = "bucket"
	teb.State = "floating"
	teb.Parent = nil
	teb.Children = make(map[string]SomaTreeBucketAttacher)
	teb.PropertyOncall = make(map[string]SomaTreeProperty)
	teb.PropertyService = make(map[string]SomaTreeProperty)
	teb.PropertySystem = make(map[string]SomaTreeProperty)
	teb.PropertyCustom = make(map[string]SomaTreeProperty)
	teb.Checks = make(map[string]Check)

	return teb
}

func (teb SomaTreeElemBucket) CloneRepository() SomaTreeRepositoryAttacher {
	cl := SomaTreeElemBucket{
		Name:        teb.Name,
		Environment: teb.Environment,
		Type:        teb.Type,
		State:       teb.State,
		Frozen:      teb.Frozen,
		Deleted:     teb.Deleted,
	}
	cl.Id, _ = uuid.FromString(teb.Id.String())
	cl.Team, _ = uuid.FromString(teb.Team.String())
	cl.Repository, _ = uuid.FromString(teb.Repository.String())

	f := make(map[string]SomaTreeBucketAttacher)
	for k, child := range teb.Children {
		f[k] = child.CloneBucket()
	}
	cl.Children = f

	pO := make(map[string]SomaTreeProperty)
	for k, prop := range teb.PropertyOncall {
		pO[k] = prop.Clone()
	}
	cl.PropertyOncall = pO

	pSv := make(map[string]SomaTreeProperty)
	for k, prop := range teb.PropertyService {
		pSv[k] = prop.Clone()
	}
	cl.PropertyService = pSv

	pSy := make(map[string]SomaTreeProperty)
	for k, prop := range teb.PropertySystem {
		pSy[k] = prop.Clone()
	}
	cl.PropertySystem = pSy

	pC := make(map[string]SomaTreeProperty)
	for k, prop := range teb.PropertyCustom {
		pC[k] = prop.Clone()
	}
	cl.PropertyCustom = pC

	cK := make(map[string]Check)
	for k, chk := range teb.Checks {
		cK[k] = chk.Clone()
	}
	cl.Checks = cK

	return &cl
}

//
// Interface: SomaTreeBuilder
func (teb *SomaTreeElemBucket) GetID() string {
	return teb.Id.String()
}

func (teb *SomaTreeElemBucket) GetName() string {
	return teb.Name
}

func (teb *SomaTreeElemBucket) GetType() string {
	return teb.Type
}

func (teb *SomaTreeElemBucket) setAction(c chan *Action) {
	teb.Action = c
}

func (teb *SomaTreeElemBucket) setActionDeep(c chan *Action) {
	teb.setAction(c)
	for ch, _ := range teb.Children {
		teb.Children[ch].setActionDeep(c)
	}
}

//
// Interface: SomaTreeBucketeer
func (teb *SomaTreeElemBucket) GetBucket() SomaTreeReceiver {
	return teb
}

func (teb *SomaTreeElemBucket) GetEnvironment() string {
	return teb.Environment
}

func (teb *SomaTreeElemBucket) GetRepository() string {
	return teb.Repository.String()
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
