package utils

import (
	"crypto/cipher"
	"crypto/aes"
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/ipfs/go-ipfs-api"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"
)


func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

//AES加密算法
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//AES解密算法
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

//随机生成AES加密秘钥
func Krand(size int, kind int) []byte {
	ikind, kinds, result := kind, [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}, make([]byte, size)
	is_all := kind > 2 || kind < 0
	rand.Seed(time.Now().UnixNano())
	for i :=0; i < size; i++ {
		if is_all { // random ikind
			ikind = rand.Intn(3)
		}
		scope, base := kinds[ikind][0], kinds[ikind][1]
		result[i] = uint8(base+rand.Intn(scope))
	}
	return result
}


//对[]byte做Md5 返回哈希字符串
func Md5(content []byte) string{

	hash := md5.Sum(content)
	hashstr := fmt.Sprintf("%x", hash)

	return hashstr
}

var sh *shell.Shell

//IPFS交互实现_上传
func UploadIPFS(fcontent []byte) (string, error) {
	sh = shell.NewShell("localhost:5001")
	hash, err := sh.Add(bytes.NewBuffer(fcontent))
	if err != nil {
		return "", err
	}
	return hash, nil
}

//IPFS交互实现_下载
func CatIPFS(hash string) ([]byte, error) {
	sh = shell.NewShell("localhost:5001")

	read, err := sh.Cat(hash)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(read)
	return body, err
}



//获取图片文件的像素点信息
func OpenImage (imgAddr string) (image.Image, error) {
	imgFile,err := os.Open(imgAddr);//打开图片文件
	if err != nil {
		return nil,err;
	}
	defer imgFile.Close();
	//根据文件的扩展名选择合适的解码方式
	var img image.Image;
	switch strings.ToLower(path.Ext(imgAddr)) {
	case ".jpg":
		img,err = jpeg.Decode(imgFile);
		if err != nil {
			return nil,err;
		}
	case ".jpeg":
		img,err = jpeg.Decode(imgFile);
		if err != nil {
			return nil,err;
		}
	case ".png":
		img,err = png.Decode(imgFile);
		if err!= nil {
			return nil,err;
		}
	default:
		return nil,errors.New("Failed to decode the image.");
	}
	return img,nil;
}


//添加水印
func WaterMarking (imgAddr string,message []byte,markedAddr string) error {
	//读取图片
	img,err := OpenImage(imgAddr);
	if err != nil {
		return err;
	}
	//获取图片大小
	bound := img.Bounds();
	//计算容量
	var capacity int = (bound.Max.X-bound.Min.Y)*(bound.Max.Y-bound.Min.Y)*3/8;
	if len(message) > capacity {
		return errors.New("Message is too long.");
	}
	newImg := image.NewNRGBA(image.Rect(bound.Min.X,bound.Min.Y,bound.Max.X,bound.Max.Y));
	var iter,i uint = 0,0;//iter是message字节切片的索引，i是指向字节内部比特位的索引
	var flag bool = true;//是否还有数据要写入图片
	//将数据写入图片
	for x:= bound.Min.X; x < bound.Max.X;x++ {
		for y := bound.Min.Y;y < bound.Max.Y;y++ {
			rt,gt,bt,at := img.At(x,y).RGBA();
			r:=uint8(rt)&0xFE;
			g:=uint8(gt)&0xFE;
			b:=uint8(bt)&0xFE;
			a:=uint8(at);
			if flag {//flag控制是嵌入数据还是复制图片
				r = r + (message[iter]>>(7-i))&0x1//将iter所指字节的第i位写入颜色的最后一位
				i++;//将i指向字节内的下一个比特
				if i>7 {//如果字节中的最后一个比特已被写入图片，则将i重置，并使iter指向下一个字节
					i = 0;
					if iter == uint(len(message)-1){//如果message中的所有字节都被写入，则继续复制图片但不在嵌入数据，否则增加iter
						flag = false;
						goto JumpOut;
					} else {
						iter++;
					}
				}
				g = g + (message[iter]>>(7-i))&0x1;
				i++;
				if i>7 {
					i = 0;
					if iter == uint(len(message)-1){
						flag = false;
						goto JumpOut;
					} else {
						iter++;
					}
				}
				b = b + (message[iter]>>(7-i))&0x1;
				i++;
				if i>7 {
					i = 0;
					if iter == uint(len(message)-1){
						flag = false;
						goto JumpOut;
					} else {
						iter++;
					}
				}
			}
		JumpOut:
			newImg.Set(x,y,color.NRGBA{R:r,G:g,B:b,A:a});
		}
	}
	markedFile,err := os.Create(markedAddr);
	if err != nil {
		return err;
	}
	png.Encode(markedFile,newImg);
	markedFile.Close();
	return nil;
}


func ReadWaterMark (imgAddr string) ([]byte,error) {
	img,err := OpenImage(imgAddr);
	if err != nil {
		return nil,err;
	}
	//获取图片大小
	bound := img.Bounds();
	var mbuffer []byte;
	var bbuffer byte = 0x0;
	var p uint = 0;
	for x:= bound.Min.X; x < bound.Max.X;x++ {
		for y := bound.Min.Y; y < bound.Max.Y; y++ {
			r,g,b,_ := img.At(x,y).RGBA();

			bbuffer += uint8(r&0x1);//从r像素获取一比特信息
			p++;
			if p>7 {//如果缓冲区中装满了一个字节，则将字节缓冲区中的数据装入切片，并清空缓冲区，重制指针
				if bbuffer == 0x0 {
					goto MessageOver;
				}
				mbuffer = append(mbuffer,bbuffer);
				bbuffer = 0x0;
				p=0;
			} else {
				bbuffer = bbuffer << 1;
			}

			bbuffer += uint8(g&0x1);//从g像素获取一比特信息
			p++;
			if p>7 {
				if bbuffer == 0x0 {
					goto MessageOver;
				}
				mbuffer = append(mbuffer,bbuffer);
				bbuffer = 0x0;
				p=0;
			} else {
				bbuffer = bbuffer << 1;
			}

			bbuffer += uint8(b&0x1);//从b像素获取一比特信息
			p++;
			if p>7 {
				if bbuffer == 0x0 {
					goto MessageOver;
				}
				mbuffer = append(mbuffer,bbuffer);
				bbuffer = 0x0;
				p=0;
			} else {
				bbuffer = bbuffer << 1;
			}
		}
	}
MessageOver:
	return mbuffer,nil;
}
