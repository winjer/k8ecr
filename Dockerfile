FROM python:3-alpine
ENV AWS_REGION eu-west-2
RUN apk add --no-cache ca-certificates 
ADD k8ecr /
ADD autodeploy.py /
RUN pip install requests 
ENV WEBHOOK https://hooks.slack.com/services/.../.../...
CMD ./autodeploy.py
