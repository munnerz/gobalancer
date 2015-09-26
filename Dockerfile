FROM scratch

ADD gobalancer /gobalancer

ENTRYPOINT ["/gobalancer"]
