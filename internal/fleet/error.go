package fleet

type FleetManagerProcessError struct {
	Inner      error
	StdoutData string
	StderrData string
}

func (e *FleetManagerProcessError) Error() string {
	return e.Inner.Error() + "\n\nStdout:\n" + e.StdoutData + "\n\nStderr:\n" + e.StderrData
}
