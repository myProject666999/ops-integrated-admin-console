import JSEncrypt from 'jsencrypt'

interface CryptoConfig {
  enabled: boolean
  publicKey: string
}

let cryptoConfig: CryptoConfig | null = null

export async function getCryptoConfig(): Promise<CryptoConfig> {
  if (cryptoConfig) {
    return cryptoConfig
  }
  
  cryptoConfig = { enabled: false, publicKey: '' }
  return cryptoConfig
}

export function clearCryptoConfig() {
  cryptoConfig = null
}

export function rsaEncrypt(data: string, publicKey: string): string {
  const encrypt = new JSEncrypt()
  encrypt.setPublicKey(publicKey)
  const encrypted = encrypt.encrypt(data)
  if (!encrypted) {
    throw new Error('RSA加密失败')
  }
  return encrypted
}

export function rsaDecrypt(data: string, publicKey: string): string {
  const decrypt = new JSEncrypt()
  decrypt.setPublicKey(publicKey)
  const decrypted = decrypt.decrypt(data)
  if (!decrypted) {
    throw new Error('RSA解密失败')
  }
  return decrypted
}

export async function encryptPassword(password: string): Promise<string> {
  const config = await getCryptoConfig()
  
  if (!config.enabled || !config.publicKey) {
    return password
  }
  
  return rsaEncrypt(password, config.publicKey)
}

export async function encryptPasswordFields<T extends Record<string, any>>(
  data: T,
  fields: string[] = ['password', 'oldPassword', 'newPassword']
): Promise<T> {
  const config = await getCryptoConfig()
  
  if (!config.enabled || !config.publicKey) {
    return data
  }
  
  const result = { ...data }
  
  for (const field of fields) {
    if (result[field] && typeof result[field] === 'string') {
      result[field] = rsaEncrypt(result[field], config.publicKey)
    }
  }
  
  return result
}

export async function isEncryptEnabled(): Promise<boolean> {
  const config = await getCryptoConfig()
  return config.enabled
}
