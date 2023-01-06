# Copyright (c) Mainflux
# SPDX-License-Identifier: Apache-2.0

#TODO: Makefile to set up list of all 

test: test0 test2 test3 test4 test5 test6 test7 test9 test10 test11 test8 test1

test0:
	echo token is $(MF_TOKEN)

test1:
	sudo docker run --rm -v /home/aryan/mainflux/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/auth.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost

test2:
	sudo docker run --rm -v /home/aryan/mainflux/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/bootstrap.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost

test3:
	sudo docker run --rm -v /home/aryan/mainflux/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/certs.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test4:
	sudo docker run --rm -v /home/aryan/mainflux/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/consumers-notifiers.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test5:
	sudo docker run --rm -v /home/aryan/mainflux/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/http.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test6:
	sudo docker run --rm -v /home/aryan/mainflux/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/provision.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test7:
	sudo docker run --rm -v /home/aryan/mainflux/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/readers.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test8:
	sudo docker run --rm -v /home/aryan/mainflux/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/things.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test9:
	sudo docker run --rm -v /home/aryan/mainflux/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/twins.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks
	
test10:
	sudo docker run --rm -v /home/aryan/mainflux/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/users.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test11:
	sudo docker run --rm -v /home/aryan/mainflux/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/websocket.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks