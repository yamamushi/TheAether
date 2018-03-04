# Dev Notes

## Adding an event type

event_types.go : has all of the actions for the various event types. 

event_loading.go : Define whether the event needs to be watched or not (read messages MUST be watched or they will hang)

event_parser.go : Here we have validation checks for events 