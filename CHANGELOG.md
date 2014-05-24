# May 24, 2014
Released v2.0 that has breaking changes in the Fallback() methods and has the following additional features:

- naming of queues and calls
- Call() to pass calls as arguments that are called while running the queue as
  alternative to pipes
- Tee(), TeeAndRun(), TeeAndFallback() to allow to pipe into multiple calls 
  and queues
- Run() and Fallback() to pass queues as arguments
- Sub() method to embed another queue
- Get(), Set(), Value(), Collect() pseudo arguments to get values in and out of the pipe

The queue library is considered feature complete and stable.
The shortcuts q library is not considered stable yet.

# Feb 02, 2014

Released v1.1 that is 100% backward compatible with v1.0 and has the following additional features:

- Logging support via the LogDebugTo() and LogErrorsTo() methods
- A new running mode called Fallback() that stops on the first non error call.
  It also allows custom error handling.
- A new error handler called PANIC that panics on the first error.
- All features are also available in the alternativ syntax package "q".
- More examples in the examples directory.

# Jan 29, 2014 

Released v1.0 with basic functionality:

- Custom error handlers
- Piping return values as input for the next function, at self defined position
- Check() method to check for non matching signatures before running the queue
- Predefined error handlers STOP and IGNORE.
- Sub package "q" with an alternative and shorter syntax.