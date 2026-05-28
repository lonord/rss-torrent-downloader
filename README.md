# rss-torrent-downloader
This project is an RSS feed-based torrent downloader that automatically fetches and processes new episodes from subscribed feeds. Upon detecting a new episode, it automatically extracts the torrent download link and submits it to the Aria2 downloader for seamless, hands-off downloading.

## Development

This project is still under development and may not be fully functional yet.

## Configuration

The downloader accepts command line flags, environment variables, and an INI config file via `-config`.

`-on-complete-script /path/to/script` configures a script or executable to run when downloads complete. The script is invoked with the completed download file paths as command line arguments.

In a config file, use:

```ini
on_complete_script=/path/to/script
```

The equivalent environment variable is `RSS_TORRENT_DL_ON_COMPLETE_SCRIPT`.

## License

MIT
