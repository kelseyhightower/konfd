FROM scratch
MAINTAINER Kelsey Hightower <kelsey.hightower@gmail.com>
ADD konfd /konfd
ENTRYPOINT ["/konfd"]
