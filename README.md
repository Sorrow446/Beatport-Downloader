# Beatport-Downloader
Beatport downloader written in Go.
![](https://i.imgur.com/T91aKEi.png)
[Windows, Linux, macOS and Android binaries](https://github.com/Sorrow446/Beatport-Downloader/releases)

# Setup
Active LINK or LINK Pro subscription required.    
Input credentials into config file.
Configure any other options if needed.
|Option|Info|
| --- | --- |
|email|Email address.
|password|Password.
|outPath|Where to download to. Path will be made if it doesn't already exist.
|albumTemplate|Album folder naming template. Vars: album, albumArtist, upc, year.
|trackTemplate|Track filename naming template. Vars: album, albumArtist, artist, bpm, genre, isrc, title, track, trackPad, trackTotal, year.
|maxCover|true = max cover size, false = 600x600.
|omitOrigMix|Omit mix type from track filenames and tags if it's an original mix.
|keepCover|true = don't delete covers from album folders.

**FFmpeg is needed to put AAC segments into MP4 containers.**    
[Windows (gpl)](https://github.com/BtbN/FFmpeg-Builds/releases)    
Linux: `sudo apt install ffmpeg`    
Termux `pkg install ffmpeg`

# Usage
Args take priority over the same config file options.

Download two albums:   
`bp_dl_x64.exe https://www.beatport.com/release/ghost-hardware-ep/63030 https://www.beatport.com/release/kindred/872666`

Download a single album and from two text files:   
`bp_dl_x64.exe https://www.beatport.com/release/ghost-hardware-ep/63030 G:\1.txt G:\2.txt`

```
 _____         _               _      ____                _           _
| __  |___ ___| |_ ___ ___ ___| |_   |    \ ___ _ _ _ ___| |___ ___ _| |___ ___
| __ -| -_| .'|  _| . | . |  _|  _|  |  |  | . | | | |   | | . | .'| . | -_|  _|
|_____|___|__,|_| |  _|___|_| |_|    |____/|___|_____|_|_|_|___|__,|___|___|_|
                  |_|

Usage: bp_dl_x64.exe [--outpath OUTPATH] [--maxcover] [--albumtemplate ALBUMTEMPLATE] [--tracktemplate TRACKTEMPLATE] URLS [URLS ...]

Positional arguments:
  URLS

Options:
  --outpath OUTPATH, -o OUTPATH
                         Where to download to. Path will be made if it doesn't already exist.
  --maxcover, -m         true = max cover size, false = 600x600.
  --albumtemplate ALBUMTEMPLATE, -a ALBUMTEMPLATE
                         Album folder naming template. Vars: album, albumArtist, upc, year.
  --tracktemplate TRACKTEMPLATE, -t TRACKTEMPLATE
                         Track filename naming template. Vars: album, albumArtist, artist, bpm, genre, isrc, title, track, trackPad, trackTotal, year.
  --help, -h             display this help and exit
  ```
  
  # Disclaimer
- I will not be responsible for how you use Beatport Downloader.    
- Beatport brand and name is the registered trademark of its respective owner.    
- Beatport Downloader has no partnership, sponsorship or endorsement with Beatport.
