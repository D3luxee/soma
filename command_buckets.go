package main

import (
	"fmt"

	"github.com/codegangsta/cli"
)

func cmdBucketCreate(c *cli.Context) {
	utl.ValidateCliArgumentCount(c, 5)
	multKeys := []string{"repository", "environment"}
	uniqKeys := []string{}

	opts := utl.ParseVariadicArguments(multKeys,
		uniqKeys,
		multKeys, // as reqKeys
		c.Args().Tail())

	_ = utl.TryGetRepositoryByUUIDOrName(opts["repository"][0])
	envs := []string{"live", "ac1", "prelive", "qa",
		"test", "dev", "default"}
	utl.ValidateStringInSlice(opts["environment"][0], envs)

	var req somaproto.ProtoRequestBucket
	req.Bucket.Name = c.Args().First()
	req.Bucket.Repository = opts["repository"][0]
	req.Bucket.Environment = opts["environment"][0]

	_ = utl.PostRequestWithBody(req, "/buckets/")
}

func cmdBucketDelete(c *cli.Context) {
	utl.ValidateCliArgumentCount(c, 3)
	utl.ValidateCliArgument(c, 2, "repository")
	repoId := utl.TryGetRepositoryByUUIDOrName(c.Args().Get(2))
	buckId := utl.TryGetBucketByUUIDOrName(
		c.Args().Get(0),
		repoId.String())
	path := fmt.Sprintf("/buckets/%s", buckId.String())

	_ = utl.DeleteRequest(path)
}

func cmdBucketRestore(c *cli.Context) {
	utl.ValidateCliArgumentCount(c, 3)
	utl.ValidateCliArgument(c, 2, "repository")
	repoId := utl.TryGetRepositoryByUUIDOrName(c.Args().Get(2))
	buckId := utl.TryGetBucketByUUIDOrName(
		c.Args().Get(0),
		repoId.String())
	path := fmt.Sprintf("/buckets/%s", buckId.String())

	var req somaproto.ProtoRequestBucket
	req.Restore = true

	_ = utl.PatchRequestWithBody(req, path)
}

func cmdBucketPurge(c *cli.Context) {
	utl.ValidateCliArgumentCount(c, 3)
	utl.ValidateCliArgument(c, 2, "repository")
	repoId := utl.TryGetRepositoryByUUIDOrName(c.Args().Get(2))
	buckId := utl.TryGetBucketByUUIDOrName(
		c.Args().Get(0),
		repoId.String())
	path := fmt.Sprintf("/buckets/%s", buckId.String())

	var req somaproto.ProtoRequestBucket
	req.Purge = true

	_ = utl.DeleteRequestWithBody(req, path)
}

func cmdBucketFreeze(c *cli.Context) {
	utl.ValidateCliArgumentCount(c, 3)
	utl.ValidateCliArgument(c, 2, "repository")
	repoId := utl.TryGetRepositoryByUUIDOrName(c.Args().Get(2))
	buckId := utl.TryGetBucketByUUIDOrName(
		c.Args().Get(0),
		repoId.String())
	path := fmt.Sprintf("/buckets/%s", buckId.String())

	var req somaproto.ProtoRequestBucket
	req.Freeze = true

	_ = utl.PatchRequestWithBody(req, path)
}

func cmdBucketThaw(c *cli.Context) {
	utl.ValidateCliArgumentCount(c, 3)
	utl.ValidateCliArgument(c, 2, "repository")
	repoId := utl.TryGetRepositoryByUUIDOrName(c.Args().Get(2))
	buckId := utl.TryGetBucketByUUIDOrName(
		c.Args().Get(0),
		repoId.String())
	path := fmt.Sprintf("/buckets/%s", buckId.String())

	var req somaproto.ProtoRequestBucket
	req.Thaw = true

	_ = utl.PatchRequestWithBody(req, path)
}

func cmdBucketRename(c *cli.Context) {
	utl.ValidateCliArgumentCount(c, 5)
	utl.ValidateCliArgument(c, 2, "to")
	utl.ValidateCliArgument(c, 4, "repository")
	repoId := utl.TryGetRepositoryByUUIDOrName(c.Args().Get(4))
	buckId := utl.TryGetBucketByUUIDOrName(
		c.Args().Get(0),
		repoId.String())
	path := fmt.Sprintf("/buckets/%s", buckId.String())

	var req somaproto.ProtoRequestBucket
	req.Bucket.Name = c.Args().Get(2)

	_ = utl.PatchRequestWithBody(req, path)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
