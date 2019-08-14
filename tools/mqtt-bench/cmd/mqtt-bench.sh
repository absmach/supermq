#!/bin/bash

i=0
echo "BEGIN TEST " >$1.out
for mtls in false true
do
	for ret in false true
	do
		for qos in 0 1 2
		do

		echo "=================================================="
		echo "MTLS: $mtls RETAIN: $ret, QOS $qos"
		echo "=================================================="
		for pub in 1 10 100
		do
			for sub in 1 10 
			do
				for message in 100 1000
				do
					if [[ $pub -eq 100 && $message -eq 1000 ]];
					then
						continue
					fi
						
					for size in 100 500 1000
					do
					let "i += 1"
					echo "=================================================================================" >> $1.out
					echo "TEST" $i - "Pub:" $pub ", Sub:" $sub ", MsgSize:" $size ", MsgPerPub:" $message    >> $1.out
					echo "=================================================================================" >> $1.out
					if [[ $mtls ]];
					then
						./mqtt-bench --channels $3 -s $size -n $message  --subs $sub --pubs $pub  -q $qos --retain=$ret -m=true -b tcps://$2:8883 --quiet=true >> $1.out
					else
						./mqtt-bench --channels $3 -s $size -n $message  --subs $sub --pubs $pub  -q $qos  --retain=$ret -b tcp://$2:1883 --quiet=true >> $1.out
					fi
					done
				done
			done
		done
		done

	done
done 