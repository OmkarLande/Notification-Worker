package entities

// ExecutionContext is the complete runtime context passed from the TaskDispatcher
// to the TaskWorker and into Provider.Execute(). It bundles the Task, its parent
// Job, and the owning App so that providers never need to query the database
// themselves during execution.
type ExecutionContext struct {
	// Task is the specific unit of work being executed.
	Task *Task

	// Job is the parent job definition loaded once by the Dispatcher per cycle.
	Job *Job

	// App is the application that owns the job, resolved from Job.AppID.
	App *App
}
