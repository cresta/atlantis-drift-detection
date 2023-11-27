FROM public.ecr.aws/docker/library/golang:1.21.4 as build

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' -o /atlantis-drift-detection ./cmd/atlantis-drift-detection/main.go

FROM public.ecr.aws/docker/library/ubuntu:23.10

RUN  apt-get update \
  && apt-get install -y wget unzip git \
  && rm -rf /var/lib/apt/lists/*


ARG TARGETPLATFORM
ARG TERRAFORM_VERSION=1.2.3
# Download terraform for linux
RUN TFARCH=${TARGETPLATFORM}; if [ \"$TFARCH\" = \"linux/arm64\" ]; then TFARCH=arm64; fi; \
  wget --quiet https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${TFARCH}.zip \
  && unzip terraform_${TERRAFORM_VERSION}_linux_${TFARCH}.zip \
  && mv terraform /usr/bin \
  && rm terraform_${TERRAFORM_VERSION}_linux_${TFARCH}.zip

 COPY --from=build /atlantis-drift-detection /atlantis-drift-detection

 ENTRYPOINT ["/atlantis-drift-detection"]
