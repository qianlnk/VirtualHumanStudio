# VirtualHumanStudio

## 音色克隆流程

1. 准备好音频素材
2. 上传音频素材 到 /data/userid/upload/目录，创建克隆任务
3. 上传音频素材 到 音色克隆服务器
文件上传接口 
curl -v  -XPOST  https://aigc-ops-test.skyengine.com.cn/v1/file/upload2path -F "attachment=@./123123.wav" -F "path=vhs"

这样在服务器上文件位置：/data/aigc-ops/vhs/123123.wav

4. 调用接口开始克隆
示例：
```
curl --location --request POST 'https://aigc-ops-test.skyengine.com.cn/v1/model/proxy/cosyvoice2-05b:8080/api/voice/clone' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model_name": "CosyVoice2-0.5B_1",
    "speaker_name": "1_leo",
    "prompt_file": "/data/aigc-ops/vhs/123123.wav",
    "prompt_text": "你咋不说你手残好拉扯他那么笨，那我问你，我手残，我技能会不会空，会不会空？我Q是锁头的吗？我QCD多少？回答我！我Q会不会空？ 嗯？你回答我，你们这些说剑魔超魔的狗。回答 我！look in my eyes, tell me why why baby why?啊！我Q会不空？我Q是顺发的吗？我Q是不是特定区域才能有伤害？我三张Q都能打满是不是啊？说话！",
    "instruct_text": ""
}'
```
4. 克隆完成后，下载音色权重文件到 /data/userid/voice/ 目录
https://aigc-ops-test.skyengine.com.cn/v1/file/view?key=model/tts_models/checkpoint/CosyVoice2-0.5B_1/spk_info/1_leo.pt