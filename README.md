## GPT4o video processing

We split the video up into frames at 1fps, stick them in a GPT4o Vision context, and then have the AI answer questions about the video.

## Example

On this video:

https://www.youtube.com/watch?v=ahvqPh8dJ8s


[![asciicast](https://asciinema.org/a/ca45P50bJAs12wbVEcG9xHguy.png)](https://asciinema.org/a/ca45P50bJAs12wbVEcG9xHguy)

### How to Use

1. **Splitting the video into frames**
   - Ensure that the video file is in the same directory.
   - Run the `split-video` script to split the video into individual frames.
     ```sh
     ./split-video <input_video> <output_directory> [fps]
     ```

2. **Interacting with the Go program**
   - Run the Go program to interact with the split frames.
     ```sh
     go run .
     ```