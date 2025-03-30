# Start the system
echo "Starting Kafka cluster..."
docker-compose up -d

# Wait for services to be fully up
echo "Waiting for services to start..."
sleep 20

# Create topics
echo "Creating topics..."
./scripts/create-topics.sh

# Start 3 consumer instances in the background
echo "Starting consumers..."
for i in {1..3}; do
  LOG_LEVEL=info go run cmd/consumer/main.go -name="consumer$i" > consumer$i.log 2>&1 &
  CONSUMER_PIDS[$i]=$!
  echo "Started consumer $i with PID ${CONSUMER_PIDS[$i]}"
done

# Start the producer in the background
echo "Starting producer..."
go run cmd/producer/main.go > producer.log 2>&1 &
PRODUCER_PID=$!
echo "Started producer with PID $PRODUCER_PID"

# Let the system run for a bit
echo "System running, letting data flow for 30 seconds..."
sleep 30

# Checking consumer logs
echo "Checking consumer logs before disruption..."
for i in {1..3}; do
  echo "=== Consumer $i logs (last 5 lines) ==="
  tail -n 5 consumer$i.log
done

# Now simulate failure by stopping one of the Kafka brokers
echo "Simulating broker failure by stopping kafka2..."
docker stop kafka2

# Wait a bit to see how the system reacts
echo "Waiting 15 seconds to observe failover behavior..."
sleep 15

# Check consumer logs again
echo "Checking consumer logs after broker failure..."
for i in {1..3}; do
  echo "=== Consumer $i logs (last 5 lines) ==="
  tail -n 5 consumer$i.log
done

# Now kill one consumer to test consumer group rebalancing
echo "Killing consumer 2 to test consumer group rebalancing..."
kill ${CONSUMER_PIDS[2]}

# Wait a bit to see how the system reacts
echo "Waiting 15 seconds to observe consumer group rebalancing..."
sleep 15

# Check remaining consumer logs
echo "Checking remaining consumer logs after consumer failure..."
for i in {1..3}; do
  if [ $i -ne 2 ]; then
    echo "=== Consumer $i logs (last 5 lines) ==="
    tail -n 5 consumer$i.log
  fi
done

# Restore the killed broker
echo "Restoring kafka2 broker..."
docker start kafka2
sleep 15

# Verify broker is back and functioning
echo "Checking topic details after broker recovery..."
docker exec kafka1 kafka-topics --describe --topic sensor-data --bootstrap-server kafka1:29092

# Clean up
echo "Cleaning up..."
kill $PRODUCER_PID
for i in {1..3}; do
  if [ $i -ne 2 ]; then  # Consumer 2 was already killed
    kill ${CONSUMER_PIDS[$i]}
  fi
done

echo "Test complete. Check the log files for detailed information."