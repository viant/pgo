FROM ubuntu:16.04
RUN apt-get update -y && apt-get install -y git libaio1 libc-bin gcc ca-certificates

ENV GO111MODULE=on
WORKDIR /
ADD pgo .
CMD ["/pgo"]
