#!/bin/bash
# Monitor tasks 3 and 4 completion status

LOG_DIR=~/"Work/taskmaster-crush-fork/.taskmaster/file-change-tracker"
CHECK_INTERVAL=300  # 5 minutes in seconds

echo "🔍 Monitoring tasks 3 and 4 for completion..."
echo "Logs: $LOG_DIR/{3,4}.log"
echo ""

while true; do
  TASK3_STATUS=$(grep -q "Status: Success" "$LOG_DIR/3.log" 2>/dev/null && echo "complete" || echo "running")
  TASK4_STATUS=$(grep -q "Status: Success" "$LOG_DIR/4.log" 2>/dev/null && echo "complete" || echo "running")
  
  TIMESTAMP=$(date "+%Y-%m-%d %H:%M:%S")
  
  if [ "$TASK3_STATUS" = "complete" ] && [ "$TASK4_STATUS" = "complete" ]; then
    echo "[$TIMESTAMP] ✅ Both tasks complete! Generating summary..."
    echo ""
    echo "==================== TASK 3 SUMMARY ===================="
    tail -100 "$LOG_DIR/3.log"
    echo ""
    echo "==================== TASK 4 SUMMARY ===================="
    tail -100 "$LOG_DIR/4.log"
    echo ""
    echo "✅ Monitoring complete. Both tasks finished successfully."
    exit 0
  else
    echo "[$TIMESTAMP] ⏳ Task 3: $TASK3_STATUS | Task 4: $TASK4_STATUS"
    sleep $CHECK_INTERVAL
  fi
done
