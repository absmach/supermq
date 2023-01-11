# Copyright (c) Mainflux
# SPDX-License-Identifier: Apache-2.0

#TODO: Makefile to set up list of all 

#! FINAL TASK HERE :-
# TODO: Keep this container running, even after the tests are over
sudo docker run --rm -v $MF_BASE_PATH/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \                        ─╯
    run -w 4 /mnt/certs.yml -H "Authorization: $MF_TOKEN" --report=./rep.gz --base-url=http://localhost

#? Results for some commands
sudo docker exec -it 408c551fb285 bash                                                                                                                                                                                         ─╯
OCI runtime exec failed: exec failed: unable to start container process: exec: "bash": executable file not found in $PATH: unknown

╰─ sudo docker exec -it 408c551fb285 ls                                                                                                                                                                                           ─╯
Error response from daemon: Container 408c551fb2854387ba2f147d0b903c3510f1444c4e4442a6f1fbb540348c50b0 is not running
#! Docker container stops within seconds of inactivity

╰─ sudo docker start 408c551fb285                                                                                                                                                                                                 ─╯
408c551fb285

╰─ sudo docker exec -it 408c551fb285 ls                                                                                                                                                                                           ─╯
README.rst      pyproject.toml  rep.gz          schemathesis


# TODO: Make this create a report in a local file
st run --base-url=http://localhost ./api/openapi/bootstrap.yml --report=./path/to/.gz-file


test: test0 test2 test3 test4 test5 test6 test7 test9 test10 test11 test8 test1 remove

test0:
	echo token is $(MF_TOKEN) base path is $(MF_BASE_PATH)

test1:
	sudo docker run --rm -v $(MF_BASE_PATH)/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/auth.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost

testlsa:
	sudo docker run --rm -v $(MF_BASE_PATH)/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
			run ls -lha

// Keep running the container, after test is complete
test2:
	sudo docker run --name=stapi -d -v $(MF_BASE_PATH)/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/bootstrap.yml -H "Authorization: $(MF_TOKEN)" --report=mnt/01.txt --base-url=http://localhost --show-errors-tracebacks \
		&& sudo docker start -d stapi
		&& sudo docker cp stapi:/mnt/01.txt $(MF_BASE_PATH)/api/reports/certs-report.txt

test3:
	sudo docker run --rm -v $(MF_BASE_PATH)/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/certs.yml -H "Authorization: $(MF_TOKEN)" --report="$(MF_BASE_PATH)/api/reports/certs-report.txt" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test4:
	sudo docker run --rm -v $(MF_BASE_PATH)/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/consumers-notifiers.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test5:
	sudo docker run --rm -v $(MF_BASE_PATH)/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/http.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test6:
	sudo docker run --rm -v $(MF_BASE_PATH)/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/provision.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test7:
	sudo docker run --rm -v $(MF_BASE_PATH)/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/readers.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test8:
	sudo docker run --rm -v $(MF_BASE_PATH)/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/things.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test9:
	sudo docker run --rm -v $(MF_BASE_PATH)/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/twins.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks
	
test10:
	sudo docker run --rm -v $(MF_BASE_PATH)/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/users.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

test11:
	sudo docker run --rm -v $(MF_BASE_PATH)/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run -w 4 /mnt/websocket.yml -H "Authorization: $(MF_TOKEN)" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

createcontainer:
	sudo docker run --name=stapi -v $(MF_BASE_PATH)/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable

stopcontainer:
	sudo docker stop stapi

remove:
	sudo docker rm stapi


# ? Storing code for trafik middleware
[http.middlewares]
  [http.middlewares.redirectComputations.redirect]
    regex = "^/computations(.*)"
    replacement = "http://${service}:9001"
    permanent = true

[http.routers]
  [http.routers.redirect]
    rule = "PathPrefix(`/computations`)"
    service = "redirect"
    middlewares = ["redirectComputations"]
    entrypoints = ["web"]
