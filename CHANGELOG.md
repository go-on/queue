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