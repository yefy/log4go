> ./logs/sniffer.log
> ./logs/sniffer_main.log

> ./logs/console.log

go run -tags=log4_debug basic_main.go > ./logs/console.log

cat ./logs/console.log|grep "info:"
cat ./logs/console.log|grep "err:"