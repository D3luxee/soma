package somatree

type SomaTreeAttacher interface {
	Attach(a AttachRequest)
	Destroy()
	Detach()

	clearParent()
	setFault(f *SomaTreeElemFault)
	setParent(p SomaTreeReceiver)
	updateFaultRecursive(f *SomaTreeElemFault)
	updateParentRecursive(p SomaTreeReceiver)
}

// implemented by: repository
type SomaTreeRootAttacher interface {
	SomaTreeAttacher
	SomaTreePropertier
	GetName() string
	attachToRoot(a AttachRequest)
}

// implemented by: buckets
type SomaTreeRepositoryAttacher interface {
	SomaTreeAttacher
	SomaTreePropertier
	GetName() string
	attachToRepository(a AttachRequest)
	CloneRepository() SomaTreeRepositoryAttacher
}

// implemented by: groups, clusters, nodes
type SomaTreeBucketAttacher interface {
	SomaTreeAttacher
	SomaTreePropertier
	GetName() string
	attachToBucket(a AttachRequest)
	CloneBucket() SomaTreeBucketAttacher
	ReAttach(a AttachRequest)
}

// implemented by: groups, clusters, nodes
type SomaTreeGroupAttacher interface {
	SomaTreeAttacher
	SomaTreePropertier
	GetName() string
	attachToGroup(a AttachRequest)
	CloneGroup() SomaTreeGroupAttacher
	ReAttach(a AttachRequest)
}

// implemented by: nodes
type SomaTreeClusterAttacher interface {
	SomaTreeAttacher
	SomaTreePropertier
	GetName() string
	attachToCluster(a AttachRequest)
	CloneCluster() SomaTreeClusterAttacher
	ReAttach(a AttachRequest)
}

type AttachRequest struct {
	Root       SomaTreeReceiver
	ParentType string
	ParentId   string
	ParentName string
	ChildType  string
	ChildName  string
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
