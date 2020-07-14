package job

type Job struct {
	Handler func(...interface{})
	Args    []interface{}
}

func NewJob(handler func(...interface{}), args ...interface{}) *Job {
	return &Job{
		Handler: handler,
		Args:    args,
	}
}
