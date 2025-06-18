package reassembler

type IssueType uint

const (
	IssueTypeUnknown IssueType = iota
	IssueTypePackageLessThanNextSeq
	IssueTypePanicOccurred
)

type Issue[S any] struct {
	Id      IssueType
	NextSeq S
	Err     error
}

func (i Issue[S]) Is(issue Issue[S]) bool {
	return i.Id == issue.Id
}

func NewIssue[S any](id IssueType, nextSeq S, err error) *Issue[S] {
	return &Issue[S]{
		Id:      id,
		NextSeq: nextSeq,
		Err:     err,
	}
}
