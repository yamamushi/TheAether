
Table of Contents
=================

   * [Events](#events)
      * [Events Command](#events-command)
      * [Fields](#fields)
      * [Event Types](#event-types)
         * [Message Events](#message-events)
            * [ReadMessage](#readmessage)
            * [TimedMessage](#timedmessage)
            * [ReadMessageChoice](#readmessagechoice)
            * [MessageChoiceTriggerEvent](#messagechoicetriggerevent)
         * [User Attribute Check Events](#user-attribute-check-events)
            * [SkillCheck](#skillcheck)
            * [HeightCheck](#heightcheck)
            * [WeightCheck](#weightcheck)
            * [AbilityCheck](#abilitycheck)
            * [ReputationCheck](#reputationcheck)
         * [User Interaction Events](#user-interaction-events)
            * [FromWallet](#fromwallet)
            * [ToWallet](#towallet)
            * [FromItem](#fromitem)
            * [ToItem](#toitem)
            * [RewardExperience](#rewardexperience)
         * [Control Flow Events](#control-flow-events)
            * [TriggerEvent](#triggerevent)
            * [TimedTriggerEvent](#timedtriggerevent)

# Events

The events system is designed to be a flexible method for defining functions within discord that can be used as standalone "read and response" type functions, functions for checking the various attributes and skills of a player, functions for checking world information such as the time of day or weather, or chained together using the scripts system to create more complex interactions with NPC's or Items (though not necessarily limited to those two).

The `events` command is used for managing events (adding, removing, modifying, etc.), and must first be enabled in a room before it can be used with the following command (see the command permissions page for more information):

`~command enable events`

## Events Command

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| add |  |  |
| remove |  |  |
| list |  |  |
| info |  |  |
| enable | | |
| disable | | |
| listenabled | | |

[Go to top of page](#table-of-contents)


## Fields

All Events share the following field types:

**Name**

_string_

A _unique_ name for the event. Attempting to save records with duplicate names will return an error.

**Description**

_string_

A description of 60 characters or less about the event.

**Type**

_string_

The type name of the event.

**TypeFlags**

_string array_

As described in the following section, each type of event has varying fields that are applicable to them.

**PrivateResponse**

_bool_

Whether or not to send a return message as a private message rather than a public one.

**Watchable**

_bool_

Whether or not this event, when triggered, should be put into the watch queue or whether it should be triggered using the passthrough data from an event before it.

i.e. If an event is triggered that is supposed to perform a skill check, that should not be a watchable event as we want it to execute immediately rather than wait for user input to proceed.

**LoadOnBoot**

_bool_

If set to true, the event will be loaded every time the bot starts up.

**Cycles**

_int_

Number of Runs. A setting of 0 is for infinite/indefinite runs (when you want to attach an event to an NPC, for example, as a general greeting).

**Data**

_string array_

As described in the following section, each type of event has varying data fields that are applicable to them. 

[Go to top of page](#table-of-contents)



## Event Types

### Message Events

#### ReadMessage

A ReadMessage event will send an automatic instant response to the user when the defined keyword is found in a message.

**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 | Keyword to trigger on |


**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | Formatted message to send |



**Example Event Definition**:

In the following example, the trigger keyword is "hello", to which a response of "Hello @username!" will be sent. It is configured to start at boot, and to cycle indefinitely:

```json
{
  "name": "ExampleReadMessage",
  "description": "Trigger a response to the word hello",
  "type": "ReadMessage",
  "typeflags": [
    "hello"
  ],
  "privateresponse": false,
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "Hello _user_!"
  ]
}
```

[Go to top of page](#table-of-contents)


#### TimedMessage

A TimedMessage event will respond to a user after the configured timeout when the defined keyword is defined in a message.

**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  Keyword to trigger on |
| 1 |  Seconds to pause for (max 300) |

**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | Formatted message to send |

**Example Event Definition**:

In the following example, the event will be triggered by the keyword "hello", after which there will be a sleep period of 30 seconds. Finally, a response of "Hello @username!" will be sent. It is configured to start at boot, and to cycle indefinitely:

```json
{
  "name": "ExampleTimedMessage",
  "description": "Trigger a timed response to the word hello",
  "type": "TimedMessage",
  "TypeFlags": [
    "hello",
    "30"
  ],
  "privateresponse": false,
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "Hello _user_!"
  ]
}
```

[Go to top of page](#table-of-contents)


#### ReadMessageChoice

The ReadMessageChoice event will respond to a user with a message that corresponds to what the defined keyword is keyed to in the data array.

    Note: This will only trigger on the **first keyword match in a message.

**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  Keyword to trigger on |
| 1 |  Second keyword to trigger on |
| 2 |  Third keyword to trigger on |
| ... |  Up to ten choices may be defined |


**Data**

    Note: The number of keywords must match the number of responses defined. Or the event will not be registered and will return an error.

| Data Field # | Description |
|-----------|-------------|
| 0 | Formatted message to send when keyword 0 is found |
| 1 | Formatted message to send when keyword 1 is found |
| 2 | Formatted message to send when keyword 2 is found |
| ... | Up to ten responses may be defined |


**Example Event Definition**:

In the following example, the keywords "hello" and "goodbye" will be responded to with "Hello @username!" and "Goodbye @username!" respectively:

```json
{
  "name": "ExampleReadMessageChoice",
  "description": "Trigger responses for hello and bye",
  "type": "ReadMessageChoice",
  "TypeFlags": [
    "hello",
    "bye"
  ],
  "privateresponse": false,
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "Hello _user_!",
    "Goodbye _user_!"
  ]
}
```

[Go to top of page](#table-of-contents)


#### MessageChoiceTriggerEvent

The MessageChoiceTriggerEvent event will trigger a keyed event when the corresponding keyword is found in a message.

    Note: This will only trigger on the first keyword match in a message.

**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  Keyword to trigger on |
| 1 |  Second keyword to trigger on |
| 2 |  Third keyword to trigger on |
| ... |  Up to ten choices may be defined |

**Data**

It is not necessary to define an eventID, but the data array length **must** match the length of the TypeFlags array. If you do not have an event yet defined to trigger, a value of "nil" can be used and updated later. If a value other than _nil_ is defined, a check will be performed to ensure that the ID is valid. 

| Data Field # | Description |
|-----------|-------------|
| 0 | ID of event to trigger (or nil) |
| 1 | ID of event to trigger (or nil) |
| 2 | ID of event to trigger (or nil) |
| ... | Up to ten events may be defined |

**Example Event Definition**:

In the following example, the keywords "sword" and "dagger" will trigger eventID "d590cbc5" and "nil" respectively:

```json
{
  "name": "ExampleMessageChoiceTriggerEvent",
  "description": "Trigger an event in response to the words sword and dagger",
  "type": "MessageChoiceTriggerEvent",
  "TypeFlags": [
    "sword",
    "dagger"
  ],
  "privateresponse": false,
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "d590cbc5",
    "nil"
  ]
}
```

[Go to top of page](#table-of-contents)

## Special Event Types

These events are intended to be used for scripting and not as general purpose events. As such, they can be defined but they cannot be enabled on a per-channel basis.

### User Attribute Check Events


#### SkillCheck

**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  ... |

**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | ... |

**Example Event Definition**:

In the following example...

```json
{
  "type": "SkillCheck",
  "TypeFlags": [
    "nil"
  ],
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "nil"
  ]
}
```

[Go to top of page](#table-of-contents)


#### HeightCheck


**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  ... |

**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | ... |

**Example Event Definition**:

In the following example...

```json
{
  "type": "HeightCheck",
  "TypeFlags": [
    "nil"
  ],
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "nil"
  ]
}
```

[Go to top of page](#table-of-contents)


#### WeightCheck


**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  ... |

**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | ... |

**Example Event Definition**:

In the following example...

```json
{
  "type": "WeightCheck",
  "TypeFlags": [
    "nil"
  ],
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "nil"
  ]
}
```

[Go to top of page](#table-of-contents)


#### AbilityCheck


**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  ... |

**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | ... |

**Example Event Definition**:

In the following example...

```json
{
  "type": "AbilityCheck",
  "TypeFlags": [
    "nil"
  ],
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "nil"
  ]
}
```

[Go to top of page](#table-of-contents)


#### ReputationCheck

**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  ... |

**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | ... |

**Example Event Definition**:

In the following example...

```json
{
  "type": "ReputationCheck",
  "TypeFlags": [
    "nil"
  ],
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "nil"
  ]
}
```

[Go to top of page](#table-of-contents)


### User Interaction Events

#### FromWallet


**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  ... |

**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | ... |

**Example Event Definition**:

In the following example...

```json
{
  "type": "FromWallet",
  "TypeFlags": [
    "nil"
  ],
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "nil"
  ]
}
```

[Go to top of page](#table-of-contents)


#### ToWallet


**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  ... |

**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | ... |

**Example Event Definition**:

In the following example...

```json
{
  "type": "ToWallet",
  "TypeFlags": [
    "nil"
  ],
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "nil"
  ]
}
```

[Go to top of page](#table-of-contents)


#### FromItem


**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  ... |

**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | ... |

**Example Event Definition**:

In the following example...

```json
{
  "type": "FromItem",
  "TypeFlags": [
    "nil"
  ],
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "nil"
  ]
}
```

[Go to top of page](#table-of-contents)


#### ToItem


**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  ... |

**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | ... |

**Example Event Definition**:

In the following example...

```json
{
  "type": "ToItem",
  "TypeFlags": [
    "nil"
  ],
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "nil"
  ]
}
```

[Go to top of page](#table-of-contents)


#### RewardExperience


**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  ... |

**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | ... |

**Example Event Definition**:

In the following example...

```json
{
  "type": "RewardExperience",
  "TypeFlags": [
    "nil"
  ],
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "nil"
  ]
}
```

[Go to top of page](#table-of-contents)


### Control Flow Events

#### TriggerEvent


**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  ... |

**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | ... |

**Example Event Definition**:

In the following example...

```json
{
  "type": "TriggerEvent",
  "TypeFlags": [
    "nil"
  ],
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "nil"
  ]
}
```

[Go to top of page](#table-of-contents)


#### TimedTriggerEvent


**TypeFlags**

| TypeFlag Field # | Description |
|-----------|-------------|
| 0 |  ... |

**Data**

| Data Field # | Description |
|-----------|-------------|
| 0 | ... |

**Example Event Definition**:

In the following example...

```json
{
  "type": "TimedTriggerEvent",
  "TypeFlags": [
    "nil"
  ],
  "loadonboot": true,
  "cycles": 0,
  "data": [
    "nil"
  ]
}
```

[Go to top of page](#table-of-contents)


