# SQL.Server

Go based HTTP handler. Responds to GET/POST/PUT/DELETE requests to specified routes by translating them into database interactions. To use the program, you must start it with an argument pointing to a configuration file. An example of the configuration JSON file is provided below:

```
{
  "Database" : {
    "UnixSocket" : "/opt/local/var/run/mysql8/mysqld.sock",
    "Username" : "my-username",
    "Password" : "my-password",
    "Database" : "my-database-name"
  },
  "Host" : "localhost",
  "Port" : "9999"
}
```

**Note**: Depends on SQL.Driver
