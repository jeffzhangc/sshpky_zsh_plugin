import hmac, base64, struct, hashlib, time

def base32_decode(base32_str):
    """解码Base32字符串"""
    b32alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
    base32_str = base32_str.upper().replace("=","");
    b32str = ""
    for char in base32_str:
        if char in b32alphabet:
            b32str += char
    n = len(b32str)
    pad = 8 - len(b32str) % 8
    if pad != 8:
        b32str += "=" * pad
    bits = ""
    for i in range(n):
        bits += format(b32alphabet.index(b32str[i]), '05b')
    bytes_out = []
    for i in range(len(bits)//8):
        b = bits[i*8:(i+1)*8]
        bytes_out.append(chr(int(b,2)))
    return base64.b32decode(b32str)

def google_authenticator_token(secret):
    """计算Google Authenticator的6位验证码"""
    # 将Base32秘钥转换为字节
    secret_bytes = base32_decode(secret)

    # 获取当前时间（秒），并转换为30秒周期的计数
    counter = int(time.time() // 30)

    # 将计数器编码为8字节的比特数组
    counter_bytes = counter.to_bytes(8, byteorder='big')

    # 使用HMAC-SHA1算法计算OTP
    hmac_sha1 = hmac.new(secret_bytes, counter_bytes, hashlib.sha1).digest()

    # 动态截取密钥的一部分以生成验证码
    offset = hmac_sha1[-1] & 0x0F
    otp = ((hmac_sha1[offset] & 0x7f) << 24) | ((hmac_sha1[offset + 1] & 0xff) << 16) | \
          ((hmac_sha1[offset + 2] & 0xff) << 8) | (hmac_sha1[offset + 3] & 0xff)

    # 将OTP模1000000得到6位验证码
    otp = otp % 1000000
    return str(otp).zfill(6)


# if __name__ == '__main__':
#     secret_key = "rq3thtbak6wpjbic4zx3i5w4x76lmluc"  # 请替换为您的16位谷歌秘钥
#     print(google_authenticator_token(str(secret_key)))  # 并未实例化
#     print(google_authenticator_token("QUWUA4A4ZN65TZUR"))  # 并未实例化
