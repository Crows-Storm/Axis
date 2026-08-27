
# Rule
Algorithm Selection:
Transport Layer Hash: SHA-256
Registration Process:
Frontend: H1 = SHA256(password)
Send: { loginId, H1 }
Backend: storedHash = H1 → Store in DB
Login Process:
Frontend: H1 = SHA256(password)
clientHash = SHA256(H1 + ":" + timestampSec + ":" + requestId)
Send: { loginId, clientHash, requestId }
Backend:
1. Retrieve storedHash from DB
2. Call the password verification function, passing in storedHash, clientHash, and requestId
3. []timestampSecs is the 10 timestamps obtained from the 5 seconds before and after the verification function call
4. Encrypt using SHA256(storedHash + ":" + timestampSec + ":" + requestId) 10 times
5. Comparison, up to 10 comparisons.

# Frontend js
insomnia pre request js scripts
```js   
const CryptoJS = require('crypto-js');

const axisFolderEnv = insomnia.parentFolders.get('Axis');
if (axisFolderEnv === undefined) {
    throw Error('Folder Axis not found');
}

const requestBody = insomnia.request.body.raw;
const bodyData = JSON.parse(requestBody);

function sha256Hex(message) {
//     return CryptoJS.SHA256(message).toString(CryptoJS.enc.Hex);
    return CryptoJS.SHA256(CryptoJS.enc.Utf8.parse(message)).toString(CryptoJS.enc.Hex);

}

if (bodyData.password) {
    const h1 = sha256Hex(bodyData.password);

    const timestampSec = Math.floor(Date.now() / 1000);

    const requestId = axisFolderEnv.environment.get('X-Request-ID')
// 		const requestId = insomnia.environment.get("X-Request-ID");

    const plain = `${h1}:${timestampSec}:${requestId}`;
    const clientHash = sha256Hex(plain);

    bodyData.password = clientHash;

    insomnia.request.body.raw = JSON.stringify(bodyData);

    // debug log
    console.log("H1:", h1);
    console.log("timestampSec:", timestampSec);
    console.log("requestId:", requestId);
    console.log("clientHash:", clientHash);
    console.log("Final Body:", bodyData);
}
```