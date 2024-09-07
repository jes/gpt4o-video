## GPT4o video processing

We split the video up into frames at 1fps, stick them in a GPT4o Vision context, and then have the AI answer questions about the video.

## Example

On this video:

https://www.youtube.com/watch?v=ahvqPh8dJ8s

[![asciicast](https://asciinema.org/a/ca45P50bJAs12wbVEcG9xHguy.svg)](https://asciinema.org/a/ca45P50bJAs12wbVEcG9xHguy)

### How to Use

1. **Interacting with the Go program**
   - Run the Go program to interact with the split frames.
     ```sh
     go run . -video foo.mp4 -fps 1
     ```

2. **Non-interactive mode**
   - Run the Go program to interact with the split frames.
     ```sh
     go run . -video foo.mp4 -fps 1 -prompt-file prompt.txt
     ```