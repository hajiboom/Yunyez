import JSEncrypt from "jsencrypt";
import request from "./request";
import { ERROR_CODES } from "./codes";
import { ca } from "element-plus/es/locales.mjs";

//缓存公钥
let cachedPublicKey = ''
//公钥缓存过期时间（后端配置，ms）
const PUBLIC_KEY_EXPIRE_TIME = 3600*1000;
let publicKeyExpireTime = 0;

/**
 * 获取公钥
 */
export const getRsaPublicKey = async () => {
    //检查缓存是否过期
    if (publicKeyExpireTime > Date.now() && cachedPublicKey) {
        return cachedPublicKey
    }

    try{
        const res = await request.get('/rsa/publicKey')
        if(res.code === ERROR_CODES.SUCCESS){
            const { publicKey, expireTime } = res.data.publicKey
            //更新缓存，过期时间为当前时间加上过期时间
            cachedPublicKey = publicKey
            publicKeyExpireTime = Date.now() + expireTime* 1000
            return publicKey
        }
        throw new Error('获取公钥失败'+res.msg)
    }
    catch(err){
        console.log('获取公钥失败',err);
        throw err//抛出错误，让调用者处理
    }
}

/**
 * 对数据进行RSA加密
 * @param {string} data - 要加密的数据
 * @returns {string} - 加密后的数据
 */
export const encryptRsa =async (data) => {
    //获取公钥
    const publicKey = await getRsaPublicKey()
    //创建JSEncrypt实例
    const encrypt = new JSEncrypt()
    //公钥要保持完整格式
    encrypt.setPublicKey(`-----BEGIN PUBLIC KEY-----\n${publicKey}\n-----END PUBLIC KEY-----`);
    //加密数据
    const encryptedData = encrypt.encrypt(data)
    if(!encryptedData){
        throw new Error('加密失败,可能是公钥格式以及明文过长')
    }
    return encryptedData
}
