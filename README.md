## Log Analyse

Fast standalone tool to help you make sense of your NGINX access logs. Still work in progress. You need make and go installed on your computer to build this project. Build by:

```sh
git clone git@github.com:jasonrogena/log-analyse.git
cd log-analyse
make deps
go build
```

You can then use the sample config (change accordingly) before running the generated binary:

```sh
cp log-analyse.toml.sample log-analyse.toml
```

### Ingest

Generates an SQLite database with extracted NGINX fields from the provided access file. Run using:

```sh
./log-analyse ingest one-off <path to log file>
```

Ingest creates the following tables in the SQLite database:

**log_file**

Column | Type | Description |
--- |--- |--- |
uuid | string | Unique identifier for the file in the database |
path | string | Path to the log file on the filesystem |
no_lines | integer | Number of files in the file when processing started |
start_time | datetime	| Time processing started on the file |
end_time | datetime | Time processing ended on the file |

**log_line**

Column | Type | Description |
--- |--- |--- |
uuid | string | Unique identifier for the line in the database |
line_no | integer | Line number in the log file |
value | string | Unprocessed log line text |
start_time | datetime	| Time processing started on the line |
log_file_uuid	| string | Unique identifier for parent log file details in log_file |

**log_field**

Column | Type | Description |
--- |--- |--- |
uuid | string | Unique identifier for the field in the database |
field_type | string | Name of NGINX field (e.g x_forwarded_for) |
value_type | string | Datatype for value of field. Can be either float or string |
value_string | string | Value of field if type is string |
value_float | float | Value of filed if type is float |
start_time | datetime | Time processing started on field |
log_line_uuid | string | Unique identifier for parent log line in log_line |
