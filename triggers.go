package main

// The purpose of a trigger is to assign a script to an action

// !! To be reviewed !!
// When a trigger expects a response from an event (non-watchable events), it will create a copy of it
// With a unique name (uuid) and the event.TriggeredEvent bool set to true (so that it is not returned on event searches)
// The data field in the new event will then contain the id of the KeyValue that the trigger is expecting to read a
// Response from. Through this mechanism, user actions can be scripted to have a variety of results.
// An example would be attaching a trigger to the "chop" action, which would fire off events checking that the user has
// An axe equipped, and various skill and attribute checks, followed by location checks and target checks to see how
// Much wood is collected by a chop (or if the action fails completely).
// By stringing together events under scripts, a large breadth of options should become available.
