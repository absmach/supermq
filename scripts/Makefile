# Copyright (c) Mainflux
# SPDX-License-Identifier: Apache-2.0

#TODO: Makefile to set up list of all 

test:
	sudo docker run -v /home/aryan/mainflux/api/openapi:/mnt --network="host" schemathesis/schemathesis:stable \
		run /mnt/bootstrap.yml -H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NzMwMTY2NjYsImlhdCI6MTY3Mjk4MDY2NiwiaXNzIjoibWFpbmZsdXguYXV0aCIsInN1YiI6ImV4YW1wbGVAZWcuY29tIiwiaXNzdWVyX2lkIjoiNzE0NTk5MmYtMzZkZi00NjE5LWE1YzQtOGJkMzg2YjI3YmE5IiwidHlwZSI6MH0.ewSQC3f1i0G0pvJp0I1xqu47Q8hcGL7aM6QsiOF318I" --base-url=http://localhost --validate-schema=true --show-errors-tracebacks

# run --base-url=http://localhost -H "Authorization: $MF_TOKEN" --validate-schema=true ./api/openapi/bootstrap.yml