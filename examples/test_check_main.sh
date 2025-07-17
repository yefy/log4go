for i in {1..10}
do
  echo "index:$i"

  > ./logs/sniffer.log
  > ./logs/sniffer_main.log

  > ./logs/console.log

  go run -tags=log4_debug check_main.go --count 100000 > ./logs/console.log

  cat ./logs/console.log|grep "info:"
  cat ./logs/console.log|grep "err:"

  cat ./logs/sniffer.log |wc -l
  cat ./logs/sniffer_main.log |wc -l
done