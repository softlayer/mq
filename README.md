# MQ-Core

---

Message Queue Core, or **MQ-Core**, is a simple HTTP-based durable message queue.

Each queue is represented as a folder on disk. Each message is stored as an individual file inside its respective queue's folder. Every operation is guaranteed to have been persisted to underlying media before a successful response is issued.

Message durability, availability, movement and locking rely on POSIX file system semantics. This allows us to deploy MQ-Core on local, distributed, or in-memory file systems.

MQ-Core has a basic notion of how it should behave when clustered. The message fetching algorithm introduces a fuzziness based on the concept of a **peer**. This fuzziness allows for significant reduction in lock contention when fetching messages at the cost of guaranteed ordering. Guaranteed ordering is only lost when peer-awareness is required, however.


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

**mq**

**mq-mover**

**mq-reaper**

## Folders

```
/new
/queues
    /queue1
    /queue2
    /queue3
    ...
/delay
/remove
```

## Message Lifecycle

TODO

## Clustering

TODO