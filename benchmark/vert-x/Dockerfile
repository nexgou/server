FROM maven:3.9-eclipse-temurin-25 AS build

WORKDIR /src
COPY pom.xml ./
COPY src ./src
RUN mvn -q -DskipTests package

FROM eclipse-temurin:25-jre
WORKDIR /app
COPY --from=build /src/target/vert-x-2.0.0.jar /app/vert-x-benchmark.jar
RUN mkdir -p /app/data
ENV PORT=3006 DB_PATH=/app/data/db.sqlite SERVICE_NAME=vert-x SERVICE_VERSION=2.0.0 LOG_LEVEL=silent JAVA_TOOL_OPTIONS="-XX:MaxRAMPercentage=75 -XX:InitialRAMPercentage=25"
EXPOSE 3006
CMD ["java", "-jar", "/app/vert-x-benchmark.jar"]
