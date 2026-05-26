FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build

WORKDIR /src
COPY asp-kestrel.csproj ./
RUN dotnet restore
COPY . ./
RUN dotnet publish -c Release -o /out --no-restore

FROM mcr.microsoft.com/dotnet/aspnet:9.0
WORKDIR /app
COPY --from=build /out ./
RUN mkdir -p /app/data
ENV PORT=3003 DB_PATH=/app/data/db.sqlite SERVICE_NAME=asp-kestrel SERVICE_VERSION=2.0.0 LOG_LEVEL=silent ASPNETCORE_URLS=http://+:3003
EXPOSE 3003
CMD ["dotnet", "asp-kestrel.dll"]
