const fs = require('node:fs');
async function main() {
  const input = JSON.parse(fs.readFileSync(0, 'utf8'));
  const location = require.resolve('miniprogram-ci', { paths: [input.modulePath, input.root] });
  const ci = require(location);
  if (typeof ci.Project !== 'function' || typeof ci.upload !== 'function') throw new Error('miniprogram-ci 模块不完整');
  if (input.check) { console.log('miniprogram-ci 本地环境检查通过；未连接微信或上传代码'); return; }
  const project = new ci.Project({
    appid: input.appId,
    type: 'miniProgram',
    projectPath: input.source,
    privateKeyPath: input.uploadKey,
    ignores: ['node_modules/**/*', '**/*.key', '**/*.pem', '**/.env*'],
  });
  await ci.upload({
    project,
    version: input.version,
    desc: input.description,
    robot: input.robot,
    setting: { useProjectConfig: true },
    onProgressUpdate: () => {},
  });
  console.log('代码已上传微信开发版本，请在微信后台核对并手动提审、发布');
}
main().then(() => process.exit(0)).catch((error) => {
  console.error(error.message);
  process.exit(1);
});
