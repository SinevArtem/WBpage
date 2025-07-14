docker exec kafka kafka-topics --create --topic wb-topic --partitions 1 --replication-factor 1 --bootstrap-server kafka:9092
docker exec kafka kafka-console-consumer --topic wb-topic --from-beginning --bootstrap-server localhost:29092
