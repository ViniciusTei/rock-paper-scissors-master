echo "Killing current process"
$pid=$(sudo lsof -i -P | grep rock | awk '{print $2}')

# Check if the PID is not empty and greater than 0
if [[ -n $pid && $pid -gt 0 ]]; then
    echo "Stopping process with PID: $pid"
    # Use kill command to stop the process
    sudo kill -9 $pid

    # Check if the kill command was successful
    if [ $? -eq 0 ]; then
        echo "Process stopped successfully."
    else
        echo "Failed to stop the process."
    fi
else
    echo "Unable to determine a valid process ID."
fi

go build -o rock-paper-scissors

# Check if the kill command was successful
if [ $? -eq 0 ]; then
  echo "Build went successfully."
  sudo nohup ./rock-paper-scissors &
  exit 0
else
  echo "Build failed."
  exit 1
fi



