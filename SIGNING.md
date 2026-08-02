# 代码签名（Code Signing）

MoveKit 当前发布的 `MoveKit.exe` 是未签名的，Windows SmartScreen 会提示
"未知发布者"。要消除该提示，需要为发布产物做 Authenticode 代码签名。

推荐使用 **Azure Trusted Signing**（微软官方服务，按签名次数计费，无需
购买 EV 证书和 USB Key）。签名流水线已在 `.github/workflows/ci.yml` 的
release 作业中搭好，配置好凭据后会在打 tag 发布时自动签名。

## 前置条件

1. 一个 Azure 订阅（[azure.microsoft.com](https://azure.microsoft.com)）
2. 创建 Trusted Signing 资源：
   - 资源组 → 创建 "Trusted Signing" 资源
   - 创建证书配置文件（推荐 "Public Trust" 类型，有效期可选）
3. 创建服务主体（Service Principal）并授权：
   - Microsoft Entra ID → App registration（或使用现有应用）
   - 将服务主体加入 Trusted Signing 账户的 "Signer" 角色
   - 记下 `Tenant ID`、`Client ID`、`Client Secret`

## 配置 GitHub Secrets / Variables

在仓库 **Settings → Secrets and variables → Actions** 中配置：

| 类型 | 名称 | 值 |
| --- | --- | --- |
| Secret | `AZURE_TENANT_ID` | Entra 租户 ID |
| Secret | `AZURE_CLIENT_ID` | 服务主体应用 ID |
| Secret | `AZURE_CLIENT_SECRET` | 服务主体客户端密钥 |
| Variable | `TRUSTED_SIGNING_ACCOUNT` | Trusted Signing 账户名 |
| Variable | `TRUSTED_SIGNING_PROFILE` | 证书配置文件名称 |

> 签名端点当前写死为美国东部（`eus.codesigning.azure.net`）。如果你的
> Trusted Signing 资源在其他区域，请修改
> `.github/workflows/ci.yml` 中签名步骤的 `endpoint`。

## 验证

配置完成后，打一个新 tag（如 `v0.3.0`）触发 release 流水线。签名步骤
成功后，在 Release 资产中下载 `MoveKit.exe`，右键 → 属性 → 数字签名，
应显示由 Microsoft Azure Code Signing 颁发的证书。

## 未配置时的行为

当 `AZURE_CLIENT_ID` 等 Secret 不存在时，CI 会**自动跳过签名步骤**，
流水线照常发布未签名产物，不影响现有发布流程。
