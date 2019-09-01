FROM node:11.1.0-alpine
RUN mkdir -p /usr/src/app
RUN apk update && apk add git make
RUN git clone https://github.com/marmelab/react-admin.git /usr/src/app/react-admin
RUN cd /usr/src/app/react-admin && make install
#RUN npm link react-admin/packages/react-admin
#RUN yarn install
RUN rm -rf /usr/src/app/react-admin/examples/simple
ADD . /usr/src/app/react-admin/examples/simple
WORKDIR /usr/src/app/react-admin/examples/simple
RUN cd /usr/src/app/react-admin && make install
EXPOSE 8080
ENTRYPOINT [ "yarn", "start" ]
CMD [ "--host=0.0.0.0", "--disable-host-check" , "--useLocalIp" ]
