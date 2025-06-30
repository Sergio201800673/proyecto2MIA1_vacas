package structs

type Journaling struct {
	OperationType [10]byte
	Path          [100]byte
	Content       [100]byte
	Date          [20]byte
	Owner         [10]byte
	Permissions   [3]byte
}
