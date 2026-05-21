count=${1:-10}

echo "count:$count"
for ((i=1; i<=count; i++)); do
  echo "index:$i"

  > ./logs/sniffer.log
  > ./logs/sniffer_main.log

  > ./logs/console.log

  go run -tags=log4_debug check_main.go --count 300000 > ./logs/console.log

  cat ./logs/console.log|grep "main_log"
  cat ./logs/console.log|grep "err:"

  if grep -q "err:" ./logs/console.log; then
    echo "found err"
    exit 1
  fi

  cat ./logs/sniffer.log |wc -l
  cat ./logs/sniffer_main.log |wc -l
done