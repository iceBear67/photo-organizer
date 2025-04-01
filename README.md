# photo-organizer
A simple tool to organize your photos by their name.

# Build

Go 1.24.1 needed.
```bash
go build main.go
```

# Usage

```aiignore
❯ go run main.go -h
Usage of ...:
  -c string
        The config file
  -h    Show Help
  -o string
        Export operations to a script. If not provided, it will flood your terminal (or pipe)
  -r string
        Test regex
  -s string
        Test String
  -w    Overwrite target if exist
  -y    Confirm all operations automatically.
```

To begin with, you need to provide a configuration like this:
```yaml
src: "/home/icybear/1drv/Camera" # Where to find photos
dst: "/home/icybear/1drv/Photos" # Where to place organized photos
patterns: # Tells the program how to resolve date information from file name.
  - "wx_camera_(?<Timestamp>\\d{13})" # Named capture group `Timestamp`, `Year`, `Month` and `Day` is supported.
  - "(?<Year>\\d{4})(?<Month>\\d{2})(?<Day>\\d{2})"
```

photo-organizer is built with POSIX file calls. For cloud drives, you should mount a VFS to use it.  
In this example, a VFS is mounted at `/home/icybear/1drv`, via rclone.  

Once the configuration is done, you can generate a script by running: `./main -c ./config.yaml -o cp.sh`, which organizes your files.   
For VFS, you can replace `cp` for `rclone copy` to make this process faster as it won't copy data back and forth.

## Test the regex
photo-organizer comes with a tool to test your patterns.

```bash
$ ./main -r "(?<Year>\d{4})(?<Month>\d{2})(?<Day>\d{2})" -s "20220201"
# 2025/04/01 17:04:18 Year : 2022
# 2025/04/01 17:04:18 Month : 02
# 2025/04/01 17:04:18 Day : 01
```

The program uses patterns from your config when `-r` is absent.
```bash
$ ./main -s "20220201" -c ./config.yaml
# 2025/04/01 17:06:39 Loading rules from ./config.yaml
# 2025/04/01 17:06:39 2 rules loaded
# 2025/04/01 17:06:39 Testing rule  0
# 2025/04/01 17:06:39 Cannot match with rule 0: no match found for given string
# 2025/04/01 17:06:39 Testing rule  1
# 2025/04/01 17:06:39 Year : 2022
# 2025/04/01 17:06:39 Month : 02
# 2025/04/01 17:06:39 Day : 01
```

This may be helpful when dealing with Go-specific regexp features.