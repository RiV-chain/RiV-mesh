تمام — خلّينا نحوّل هذا الـREADME إلى إصدار أوّل نظيف “تجريبي” مع تصحيح الروابط والترخيص وصياغة واضحة، + نجهّز لك قالب إصدار (Release) وخطوات CI ترفع الأصول تلقائيًا. تحت عندك **ملف README مُصحَّح جاهز للصق** + **خطة إصدار أول** + **قالب GitHub Actions**.

---

# 1) مراجعة سريعة لأهم التعديلات المطلوبة

* **الهوية:** غيّر كل ما يشير إلى *RiV Chain / RiV-mesh / v6Space* ليشير إلى **Uqda / Uqda-Core** فقط (إلا إذا تقصد ذكر أصل المشروع في فقرة “الاعتمادات”).
* **الروابط:** وجّه الروابط إلى مستودعاتك (org **Uqda**).
  مثال: بدل `riv-chain.github.io/Uqda-Core/` استخدم صفحة GitHub Pages الخاصة بك أو احذف الروابط مؤقتًا.
* **الترخيص (مهم):** فقرة المتطلبات فيها التباس. **LGPLv3 لا تُجبِر تطبيقك على فتح مصدره**. الالتزام يقع على **المكتبة المعدَّلة نفسها**، مع السماح للمستخدم بإعادة الربط (relinking). صحّح الصياغة.
* **الأصول/الإصدارات:** إبدأ بالثنائيات (EXE/ELF/Mach-O) أولًا؛ المُثبّتات (MSI/DEB/PKG) تضيفها لاحقًا.

---

# 2) نسخة README مختصرة وصحيحة (جاهزة للصق)

> غيّر الروابط التي عليها تعليق `TODO` لاحقًا لما تجهّز صفحاتك.

````md
# Uqda-Core — Decentralized IPv6 Mesh Network

Uqda-Core is an end-to-end encrypted IPv6 mesh overlay. It’s lightweight, self-arranging, and works across Linux, Windows, macOS, BSDs, and embedded routers. Any IPv6-capable app can communicate securely with other nodes over Uqda-Core.

> **Status:** Experimental preview (first community release). Interfaces and config may change.

## Features
- Encrypted IPv6 mesh overlay (TUN/TAP)
- Auto-configuration mode and static config mode
- Cross-platform builds
- IoT-friendly footprint

## Getting Started

### Build from source
Requirements: Go 1.19+
```bash
git clone https://github.com/Uqda/Core.git
cd Core
go mod tidy
go build -o bin/mesh ./cmd/mesh
````

### Generate a config

```bash
./mesh -genconf > ./mesh.conf        # HJSON with comments
# or JSON:
./mesh -genconf -json > ./mesh.conf
```

### Run

```bash
sudo ./mesh -useconffile ./mesh.conf
# or:
sudo ./mesh -autoconf
```

> On Windows: run `mesh.exe` as Administrator, ensure IPv6 is enabled, and install WireGuard for Windows (provides Wintun) or place `wintun.dll` next to `mesh.exe`.

## Downloads

Prebuilt binaries and installers are published on the **Releases** page:

* Windows: `mesh-<version>-x64.exe` (portable) and/or `mesh-<version>-x64.msi` *(coming soon)*
* Linux/macOS: portable binaries *(packages coming soon)*

👉 Releases: **[https://github.com/Uqda/Core/releases](https://github.com/Uqda/Core/releases)**

## Documentation

* Quickstart (README)
* Configuration reference *(TODO: link to docs once published)*
* FAQ *(TODO)*
* Changelog: [CHANGELOG.md](./CHANGELOG.md)

## Known Issues

1. **Windows / Wintun**

   * Error: `Unable to load wintun.dll`
   * Fix: Install “WireGuard for Windows” (installs Wintun) or copy `wintun.dll` beside `mesh.exe`. Run as Administrator and ensure IPv6 is enabled system-wide.

2. **No IPv6 support**

   * Errors around `address family not supported` / TUN not supported. Enable IPv6 or use a platform with TUN support.

3. **Docker SCTP conflicts**

   * If logs spam `Connected/Disconnected SCTP`, try stopping Docker or adjusting bind interfaces/ports.

## License (LGPLv3)

Uqda-Core is released under **LGPLv3**.

* You may use Uqda-Core in proprietary or open applications.
* If you **modify and distribute Uqda-Core itself (the library/program)**, you must provide your modifications under LGPLv3.
* If you **link** Uqda-Core, you must not prevent users from relinking/replacing the LGPL component (e.g., provide object files or use dynamic linking as appropriate).
* Full text: [LICENSE](./LICENSE)

> Trademarks are property of their respective owners.

## Credits

This project builds on the ecosystem of IPv6 mesh networking and the community around it. We thank all contributors and upstream projects that inspired the work.
