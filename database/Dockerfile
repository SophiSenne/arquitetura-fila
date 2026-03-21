FROM postgres:16

ENV POSTGRES_DB=database
ENV POSTGRES_USER=admin
ENV POSTGRES_PASSWORD=admin

COPY ./init /docker-entrypoint-initdb.d

EXPOSE 5432