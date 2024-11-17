# gopload

**gopload** is a simple file upload service designed to be used via the command line interface (CLI). It allows users to upload files to a server where they can be stored and later retrieved. The service is built with Go and utilizes the Gin web framework for handling HTTP requests.

## Features

- **File Upload**: Upload files via HTTP PUT and POST requests.
- **File Retrieval**: Retrieve uploaded files via HTTP GET requests.
- **Automatic Cleanup**: Automatically deletes files after a specified number of days.
- **Configurable Storage**: Store files in a directory of your choice.
- **Debug Mode**: Enable debug mode for more verbose logging.
- **Cron Job**: Uses a cron job to periodically clean up old files.

## Installation

To install `gopload`, you need to have Go installed on your system. Run the following command:

```bash
go get github.com/phpgao/gopload
```

## Configuration

gopload can be configured using command line flags or environment variables. Here are the available flags:

    --listen or -b: The address to bind the server to (default: :8088).
    --dir: The directory where files will be stored (default: none).
    --debug: Enable debug mode (default: false).
    --length or -l: The length of the random directory name (default: 7).
    --max or -m: The maximum file size in MB (default: 100).
    --expire or -e: The number of days after which files are deleted (default: 3).
## Usage

To start the gopload service, run the following command:

```bash
gopload --listen :8088 --dir /path/to/storage --debug
```
Replace /path/to/storage with the directory where you want to store the uploaded files.

## Uploading Files

To upload a file, use the PUT or POST method with the filename as a path parameter:

```bash
curl -X PUT -F "file=@path/to/your/file" http://localhost:8088/yourfile.txt

curl -d "file=@path/to/your/file" http://localhost:8088/yourfile.txt
```

## Retrieving Files

To retrieve a file, use the GET method with the path and filename as path parameters:

```bash
curl http://localhost:8088/path/to/storage/yourfile.txt -O
wget http://localhost:8088/path/to/storage/yourfile.txt
```

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.

## License

gopload is released under the Apache License Version 2.0.