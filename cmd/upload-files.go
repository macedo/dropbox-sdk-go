package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/macedo/dropbox-sdk-go/dropbox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var uploadFilesCmd = &cobra.Command{
	Use:   "upload-files",
	Short: "Upload files to dropbox (do not use this to upload a file larger than 150 MiB)",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("no files to upload")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return uploadFilesAction(os.Stdout, args)
	},
}

type ErrorCtx struct {
	Error    error
	Filename string
}

func uploadFilesAction(w io.Writer, filenames []string) error {
	resCh := make(chan *dropbox.FilesUploadOutput)
	errCh := make(chan *ErrorCtx)
	doneCh := make(chan struct{})

	cli, err := dropbox.New(
		dropbox.WithCredentialsPath(viper.ConfigFileUsed()),
	)
	if err != nil {
		return fmt.Errorf("failed to create new dropbox client: %v", err)
	}

	wg := sync.WaitGroup{}

	for _, fname := range filenames {
		wg.Add(1)

		go func(fname string) {
			defer wg.Done()

			f, err := os.Open(fname)
			if err != nil {
				errCh <- &ErrorCtx{fmt.Errorf("failed to open file: %v", err), fname}
				return
			}

			path := viper.GetString("dropbox-path") + fname

			out, err := cli.FilesUpload(&dropbox.FilesUploadInput{
				Body: f,
				UploadArg: &dropbox.UploadArg{
					Mode: "overwrite",
					Mute: false,
					Path: path,
				},
			})
			if err != nil {
				errCh <- &ErrorCtx{fmt.Errorf("failed to upload file %s: %v", path, err), fname}
				return
			}

			resCh <- out
		}(fname)
	}

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	for {
		select {
		case err := <-errCh:
			fmt.Printf("Failed to upload %s: %v\n", err.Filename, err.Error)
		case out := <-resCh:
			fmt.Fprintf(w, "File uploaded => %s\n", out.PathDisplay)
		case <-doneCh:
			return nil
		}
	}
}

func init() {
	uploadFilesCmd.PersistentFlags().String("dropbox-path", "/", "Path in the user's Dropbox to save the file.")
	viper.BindPFlag("dropbox-path", uploadFilesCmd.PersistentFlags().Lookup("dropbox-path"))

	rootCmd.AddCommand(uploadFilesCmd)
}
