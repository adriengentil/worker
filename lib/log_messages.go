package lib

const (
	noLogOutputMessage = `

No output has been received in the last %d minutes, this potentially indicates a stalled build or something wrong with the build itself.

The build has been terminated.

`
	logTooLongMessage = `

The log length has exceeded the limit of %d MiB (this usually means that the test suite is raising the same exception over and over).

The build has been terminated.

`
	stalledBuildMessage = `

We're sorry but your test run exceeded %d minutes.

One possible solution is to split up your test run.

`
	connectionErrorMessage = `

We're sorry, but there was an error with the connection to the VM.

Your job will be requeued shortly.

`

	vmCreationErrorMessage = `

We're sorry, but there was an error starting a new VM.

Your job will be requeued shortly.

`
	cancelledJobMessage = `

Done: Job Cancelled

`
)