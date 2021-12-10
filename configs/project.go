package configs

const (
	DONE   = 1
	UNDONE = 2
)

var TaskStatusMap = map[int]string{
	DONE:   "Done",
	UNDONE: "Undone",
}
