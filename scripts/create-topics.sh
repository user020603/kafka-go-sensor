# Create the first topic with 3 partitions and replication factor 2
docker exec kafka1 kafka-topics --create --topic sensor-data \
  --bootstrap-server kafka1:29092 \
  --partitions 3 \
  --replication-factor 2

# Create the second topic with 3 partitions and replication factor 2
docker exec kafka1 kafka-topics --create --topic system-logs \
  --bootstrap-server kafka1:29092 \
  --partitions 3 \
  --replication-factor 2

# Check if topics were created
docker exec kafka1 kafka-topics --list --bootstrap-server kafka1:29092

# Describe topics to see partition and replication details
docker exec kafka1 kafka-topics --describe --topic sensor-data --bootstrap-server kafka1:29092
docker exec kafka1 kafka-topics --describe --topic system-logs --bootstrap-server kafka1:29092