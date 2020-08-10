#!/bin/bash

TEST_INTERVAL=600

while true
do
  echo "Begin test at $(date)"

  bats startup_time.bats
  if [ $? -ne 0 ]; then
      curl -X POST "https://oapi.dingtalk.com/robot/send?access_token=$TOKEN" -H 'cache-control: no-cache' -H 'content-type: application/json' -d '{
        "msgtype": "text",
        "text": {
            "content": "terway startup time test failed!"
          }
        }'
    else
      echo "Test succeed at $(date)"
    fi

    sleep $TEST_INTERVAL
done
