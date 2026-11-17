from confluent_kafka import Consumer, KafkaError
from miko.core.logging.logger import get_logger

okitsune = get_logger("kafka")

consumer = Consumer(
    {
        "bootstrap.servers": "localhost:9092",
        "group.id": "miko",
        "enable.auto.commit": False,
    }
)

consumer.subscribe(["cases.opened"])

while True:
    message = consumer.poll(1.0)

    if message is None:
        continue

    if message.error():
        if message.error().code() == KafkaError._PARTITION_EOF:
            okitsune.debug(
                "Reached end of partition | topic=%s | partition=%s",
                message.topic(),
                message.partition(),
            )
            continue
        okitsune.error("Kafka error: %s", message.error())
        continue
    
    raw_event = message.value()

    try:
        event = CaseOpened()
        event.ParseFromString(raw_event)

    except Exception as e:
        okitsune.error("Failed to decode CaseOpened: %s", e)

        consumer.commit(message=message, asynchronous=False)
        continue

    consumer.commit(message=message, asynchronous=False)
