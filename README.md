# MQ

Message Queue, or **MQ**, is a simple HTTP-based durable message queue.

Each queue is represented as a folder on disk. Each message is stored as an individual file inside its respective queue's folder. Every operation is guaranteed to have been persisted to underlying media before a successful response is issued.

Message durability, availability, movement and locking rely on POSIX file system semantics. This allows us to deploy MQ on local, distributed, or in-memory file systems.

MQ has a basic notion of how it should behave when clustered. The message fetching algorithm introduces a fuzziness based on the concept of a **peer**. This fuzziness allows for significant reduction in lock contention when fetching messages at the cost of guaranteed ordering. Guaranteed ordering is only lost when peer-awareness is required, however.

## Endpoints

There are only 6 endpoints in total. Additional functionality required should be implemented by your application.

#### Queues

**Fetch a queue.**

```
GET         /queue001
```

If the queue exists, an HTTP status code of **200** (OK) will be returned.

**Create a queue.**

```
PUT         /queue001
```

Only the path and method are significant, the body of the request is discarded.

**Delete a queue.**

```
DELETE      /queue001
```

This will delete the queue and all its messages.

#### Messages

**Fetch a message from a queue.**

```
GET         /queue001/messages
```

The body of the response will be the message content, and the **X-Message-Id** header of the response will contain the message ID.

If a message cannot be fetched, an HTTP status code of **204** (No Content) will be returned.

**Add a message to a queue.**

```
POST        /queue001/messages
```

The body of the request will be the message content, and the **X-Message-Id** header of the response will contain the message ID.

If the request body cannot be read, an HTTP status code of **400** (Bad Request) will be returned.

If the message cannot be guaranteed as stored, an HTTP status code of **503** (Service Unavailable) will be returned. The response will also include the header **Retry-After** with an integer value of how many seconds to wait before reissuing the request.

**Delete a message from a queue.**

```
DELETE      /queue001/messages/4ad814ab-213e-11e3-a9a3-0025904f6e08
```

The ID provided should be the same ID returned from fetching or adding a message to the queue.

## Daemons

MQ is comprised of 3 daemons. In order for the system to remain available, only the **mq** daemon must be running and responsive.

#### mq

*Handles all requests for the HTTP endpoints.*

```
--workers=8
```

The number of internal worker pairs to spawn. For each increment of this value, one message fetching worker and one message saving worker will be started. Defaults to **8**.

```
--peers=0
```

The number of peers to take into consideration when fetching messages. Defaults to **0**.

```
--root=/tmp/mq
```

The directory in which the folder structure will be created. Must be writable. Defaults to **/tmp/mq**.

#### mq-mover

*Used for moving files between source to destination directories, optionally with a delay per file.*

```
--source
```

The directory from which files will be taken. Must be readable. No default, this is required.

```
--destination
```

The directory to which files will be delivered. Must be writable. No default, this is required.

```
--delay=0
```

The time, in seconds, to wait after taking and before delivering a file to its destination. Defaults to **0**.

#### mq-reaper

*Removes files from a source directory.*

```
--source=/tmp/mq/remove
```

The directory from which files will be unlinked. Must be writable. Defaults to **/tmp/mq/remove**.

## Folders

The first time the **mq** daemon is started, this folder structure is created under the specified root directory.

```
/new
/queues
    /queue001
    /queue002
    /queue003
    ...
/delay
/remove
```

**/new**: Contains inbound messages.

**/queues**: Contains one folder per queue.

**/queues/queue00{1,2,3,...}**: Each contains one file per message for the queue it represents.

**/delay**: Contains message files recently fetched.

**/remove**: Contains message files to be removed.

## Message Lifecycle

#### Message Creation

An inbound message is first written to the **new** folder as an individual file. The file name takes the form of **queue001:id**. The content of the file are the bytes posted as the body of the request to create the message. An affirmative response is only returned if the file is successfully flushed to stable media, otherwise we considered it failed and attempt to clean up whatever partial write may have occurred.

#### Message Delivery

Once the message has been fully written to the **new** folder, it is moved into the folder indicated by its filename prefix (everything occurring before the colon, ':'). In this case, **queue001**. Its new filename is only the **id** inside the target queue's folder. So, **/new/queue001:id** is moved to **/queues/queue001/id**. This operation guarantees atomic delivery, as a partial file is never revealed to inbound requests for messages.

#### Message Fetching

All requests for messages are serialized per-queue.

In the event **peers** is larger than zero, the fetching behavior changes in an attempt to reduce message duplication.

TODO: Describe the fetching algorithm.

#### Message Delay && Re-Delivery

During the fetching of a message, it is palced into the **delay** folder using the same file name it had at the time of its creation. So, **/queues/queue001/id** is moved to **/delay/queue001:id**. Upon arrival in the delay folder, a timer is applied to the message. Once this timer expires, the message is re-delivered to its queue. So, **/delay/queue001:id** is moved to **/queues/queue001/id**.

#### Message Removal

At any point after creation and before removal, a message can be removed. Attempts to move **/new/queue001:id**, **/queues/queue001/id**, and **/delay/queue001:id** to **/remove/queue001:id** are made in sequence. The first movement to succeed is considered a successful removal and ends the sequence of attempts.

#### Message Destruction

Once the message arrives in the **remove** folder, at some point in the future it will be permanently removed from stable media. It should be assumed this action is instant, but that is not guaranteed.

## Message Movement

A message is only ever written once. A message is only ever unlinked once. All delivery, re-delivery, delay and removal activity is achieved through file system move operations. This should be taken into consideration when dealing with distributed file systems, partition boundaries, and file system journals.

## License

Copyright © 2013 SoftLayer, an IBM Company. Our code and documentation is licensed under the MIT license. By contributing your code, you agree to license your contribution under the terms of this license.

## Acknowledgements

The UUIDv1 algorithm is based on Christoph Hack's work for the UUID module in [tux21b/gocql](https://github.com/tux21b/gocql). The original work is BSD-licensed.
