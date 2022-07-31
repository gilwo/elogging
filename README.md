# elogging - an enhanced logging package based on golang log package 

## included features
* scoped logging - unique log object per scope - (scope is not unique)
* leved logging - diffrentiate between logging messages based on logic
* simple text level - `error`, `warn`, `info`, `verbose`, `trace` and `disabled`
  - set explictly
  - cycle the levels
  - use simple text or use defines from the library
* simple on/off for all log objects
* all log objects accessiable from the library
* control output for each log object 
* change default setting for log creation
* globally set log level

