package utilsw

import (
	"bytes"
	"io"
)

// BytesFilter 接口定义了一个方法 Accept，用于过滤字节切片。
// 这个接口的实现需要提供一个 Accept 方法，该方法接收一个字节切片 buf，
// 并返回一个经过过滤的字节切片以及一个布尔值，指示是否应该继续处理返回的缓冲区。
//
// Accept 方法的参数:
//
//	buf []byte: 待过滤的字节切片。
//
// Accept 方法的返回值:
//
//	[]byte: 过滤后的字节切片。
//	bool: 一个布尔值，指示是否应该继续处理返回的缓冲区。
type BytesFilter interface {
	Accept(buf []byte) ([]byte, bool)
}

// FilterReader 返回一个经过 ByteFilter 过滤的 io.Reader。
// 这个函数通过读取原始的 io.Reader，使用 ByteFilter 进行数据过滤，
// 并将过滤后的数据写入到一个管道中，最终返回这个管道的读取端。
// 参数:
//
//	src io.Reader: 原始的输入流。
//	filter ByteFilter: 用于过滤数据的 ByteFilter 实例。
//
// 返回值:
//
//	io.Reader: 返回一个经过数据过滤的 io.Reader，通常用于按特定条件过滤数据。
func FilterReader(src io.Reader, filter BytesFilter) io.Reader {
	// 创建一个管道，用于后续返回过滤后的数据读取端。
	pr, pw := io.Pipe()

	// 启动一个 goroutine 从 src 读取数据，进行过滤，并将结果写入 pw。
	go func() {
		// 确保在函数退出时关闭管道的写入端。
		defer pw.Close()
		defer pr.Close()

		// 创建一个缓冲区用于读取数据。
		b := make([]byte, 8192*1024)
		// 创建一个缓冲区，用于暂存需要延迟处理的数据。
		var buf bytes.Buffer
		for {
			// 从 src 读取数据。
			n, err := src.Read(b)
			if n > 0 {
				// 当前读取的数据。
				curr := b[:n]

				// 如果缓冲区中有数据，将当前读取的数据追加到缓冲区中。
				total := curr
				if buf.Len() > 0 {
					total = append(buf.Bytes(), curr...)
				}

				// 使用 ByteFilter 对当前数据进行过滤。
				accept, needHold := filter.Accept(total)

				if needHold {
					// 如果需要延迟处理，将数据存入缓冲区，并继续下一次循环。
					buf.Write(curr)
					continue
				} else {
					// 如果不需要延迟处理，重置缓冲区。
					if buf.Len() > 0 {
						buf.Reset()
					}

					// 将过滤后的数据写入管道。
					pw.Write(accept)
				}
			}

			// 如果遇到 EOF，跳出循环。
			if err == io.EOF {
				break
			}

			// 如果发生其他错误，关闭管道的写入端，并跳出循环。
			if err != nil {
				pw.CloseWithError(err)
				break
			}
		}

		// 如果缓冲区中仍有数据，进行最后一次过滤并写入管道。
		if buf.Len() > 0 {
			accept, _ := filter.Accept(buf.Bytes())
			// fmt.Println("final", buf.String())
			pw.Write(accept)
		}
	}()

	// 返回管道的读取端。
	return pr
}
