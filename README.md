# VirtualHumanStudio

https://huggingface.co/camenduru/FLUX.1-dev/resolve/main/ae.sft -d /content/ComfyUI/models/vae -o ae.sft
https://huggingface.co/camenduru/FLUX.1-dev/resolve/main/clip_l.safetensors -d /content/ComfyUI/models/clip -o clip_l.safetensors

https://huggingface.co/camenduru/FLUX.1-dev/resolve/main/t5xxl_fp8_e4m3fn.safetensors  -d /content/ComfyUI/models/clip

https://huggingface.co/camenduru/FLUX.1-dev/resolve/main/flux1-redux-dev.safetensors?download=true √
    
https://huggingface.co/funnewsr/sigclip_vision_patch14_384/resolve/main/sigclip_vision_patch14_384.safetensors

https://huggingface.co/ersdd/temp/resolve/f0ca3a4d938903440dfe2ff47b68ad184d338e11/F.1-Fill-fp16_Inpaint%26Outpaint_1.0.safetensors



comfyui物品替换模块，技术实现流程，页面提交物品白底图、模特图、模特蒙版图，这些图片上传到comfyui服务端，再结合内置工作流提交画图任务到comfyui，异步查询任务进度，完成后下载生成的图片，更新到物品替换任务，查询任务详情展示输入输出图片，帮我实现下这套流程,要求页面提供物品白底图跟模特图上传入口，打开图片可以进入图片跟图片画上蒙版，保存蒙版图片，点击提交按钮，提交到后台服务，后台服务调用comfyui接口提交任务，异步查询任务进度，完成后下载生成的图片，更新到物品替换任务，查询任务详情展示输入输出图片
comfyui图片上传接口：
curl --location --request POST 'http://192.168.2.184:8188/upload/image' \
--form 'image=@"C:\\Users\\ASUS\\Pictures\\00001.png"'
comfyui蒙版图片上传接口：
curl --location --request POST 'http://192.168.2.184:8188/upload/mask' \
--form 'image=@""' \
--form 'type="input"' \
--form 'subfolder=""' \
--form 'original_ref="{\"filename\":\"af1176bc61309e4b5560a3d2e881a8b5c1914593ff36366a86f739e67ad10027.png\",\"type\":\"input\"}"'
comfyui提交绘图任务接口：
curl --location --request POST 'http://192.168.2.184:8188/prompt' \
--header 'Content-Type: application/json' \
--data-raw '{
    "client_id": "test003",
    "prompt": {
        "3": {
            "inputs": {
                "seed": 1095203137526772,
                "steps": 20,
                "cfg": 8,
                "sampler_name": "euler",
                "scheduler": "normal",
                "denoise": 1,
                "model": [
                    "4",
                    0
                ],
                "positive": [
                    "6",
                    0
                ],
                "negative": [
                    "7",
                    0
                ],
                "latent_image": [
                    "5",
                    0
                ]
            },
            "class_type": "KSampler",
            "_meta": {
                "title": "K采样器"
            }
        },
        "4": {
            "inputs": {
                "ckpt_name": "ghostxl_v10BakedVAE.safetensors"
            },
            "class_type": "CheckpointLoaderSimple",
            "_meta": {
                "title": "Checkpoint加载器(简易)"
            }
        },
        "5": {
            "inputs": {
                "width": 1024,
                "height": 1024,
                "batch_size": 1
            },
            "class_type": "EmptyLatentImage",
            "_meta": {
                "title": "空Latent"
            }
        },
        "6": {
            "inputs": {
                "text": "a beautiful girl",
                "clip": [
                    "4",
                    1
                ]
            },
            "class_type": "CLIPTextEncode",
            "_meta": {
                "title": "CLIP文本编码器"
            }
        },
        "7": {
            "inputs": {
                "text": "text, watermark",
                "clip": [
                    "4",
                    1
                ]
            },
            "class_type": "CLIPTextEncode",
            "_meta": {
                "title": "CLIP文本编码器"
            }
        },
        "8": {
            "inputs": {
                "samples": [
                    "3",
                    0
                ],
                "vae": [
                    "4",
                    2
                ]
            },
            "class_type": "VAEDecode",
            "_meta": {
                "title": "VAE解码"
            }
        },
        "9": {
            "inputs": {
                "filename_prefix": "ComfyUI",
                "images": [
                    "8",
                    0
                ]
            },
            "class_type": "SaveImage",
            "_meta": {
                "title": "保存图像"
            }
        }
    }
}'
comfyui查询任务进度接口：
curl --location --request GET 'http://192.168.2.184:8188/history/8278b4e6-656a-4dbd-8ec9-4a0784babac7'
comfyui取消当前任务接口：
curl --location --request POST 'http://192.168.2.184:8188/interrupt'


curl 'https://aigc-ops-test.skyengine.com.cn/v1/model/comfyui/lnk-comfyui/api/prompt' \
  -H 'accept: */*' \
  -H 'accept-language: zh-CN,zh;q=0.9' \
  -H 'cache-control: no-cache' \
  -H 'comfy-user;' \
  -H 'content-type: application/json' \
  -b 'aigc_web_platform_token_stage=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDM0OTc5NzMsIm5iZiI6MTc0Mjg5MzE3MywiaWF0IjoxNzQyODkzMTczLCJ1aWQiOjQxMTIsInVzZXJfaWQiOiJUNDExMiIsInVzZXJfbmFtZSI6IuiwouaMr-WutiIsImRlcHQiOiLnoJTlj5HkuK3lv4Mt5pWw5o2u5pm66IO957q_LeWqkuS9k-S4jkFJR0Pnu4Qt5bel56iL57uEIiwibmFtZSI6IuiwouaMr-WutiIsImRlcGFydG1lbnQiOiLnoJTlj5HkuK3lv4Mt5pWw5o2u5pm66IO957q_LeWqkuS9k-S4jkFJR0Pnu4Qt5bel56iL57uEIiwiZGVwdDEiOiLnoJTlj5HkuK3lv4MiLCJkZXB0MiI6IuaVsOaNruaZuuiDvee6vyIsImVtYWlsIjoieGllemhlbmppYUA1MnR0LmNvbSIsInJvbGUiOjEsInZlcnNpb24iOjF9.7R0v5CstEkwiRQGiPd5x1qY71H9gABCgDtw38OAcEYI' \
  -H 'origin: https://aigc-ops-test.skyengine.com.cn' \
  -H 'pragma: no-cache' \
  -H 'priority: u=1, i' \
  -H 'referer: https://aigc-ops-test.skyengine.com.cn/v1/model/comfyui/lnk-comfyui/' \
  -H 'sec-ch-ua: "Chromium";v="134", "Not:A-Brand";v="24", "Google Chrome";v="134"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "macOS"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-origin' \
  -H 'user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36' \
  --data-raw '{"client_id":"ab1aaaa82e6b4a6487588252cc6a95f8","prompt":{"10":{"inputs":{"image":"clipspace/clipspace-mask-1961925.8999999762.png [input]"},"class_type":"LoadImage","_meta":{"title":"加载图像"}},"35":{"inputs":{"seed":228408831139036,"steps":30,"cfg":1,"sampler_name":"euler","scheduler":"normal","denoise":1,"model":["46",0],"positive":["86",0],"negative":["45",1],"latent_image":["45",2]},"class_type":"KSampler","_meta":{"title":"K采样器"}},"36":{"inputs":{"text":"","clip":["44",0]},"class_type":"CLIPTextEncode","_meta":{"title":"CLIP文本编码"}},"37":{"inputs":{"samples":["35",0],"vae":["43",0]},"class_type":"VAEDecode","_meta":{"title":"VAE解码"}},"40":{"inputs":{"text":"The man is wearing glasses","clip":["44",0]},"class_type":"CLIPTextEncode","_meta":{"title":"CLIP文本编码"}},"41":{"inputs":{"guidance":30,"conditioning":["40",0]},"class_type":"FluxGuidance","_meta":{"title":"Flux引导"}},"42":{"inputs":{"unet_name":"F.1-Fill-fp16_Inpaint&Outpaint_1.0.safetensors","weight_dtype":"fp8_e4m3fn_fast"},"class_type":"UNETLoader","_meta":{"title":"UNet加载器"}},"43":{"inputs":{"vae_name":"ae.sft"},"class_type":"VAELoader","_meta":{"title":"加载VAE"}},"44":{"inputs":{"clip_name1":"clip_l.safetensors","clip_name2":"t5xxl_fp8_e4m3fn.safetensors","type":"flux","device":"default"},"class_type":"DualCLIPLoader","_meta":{"title":"双CLIP加载器"}},"45":{"inputs":{"noise_mask":false,"positive":["41",0],"negative":["36",0],"vae":["103",0],"pixels":["96",0],"mask":["95",0]},"class_type":"InpaintModelConditioning","_meta":{"title":"内补模型条件"}},"46":{"inputs":{"model":["42",0]},"class_type":"DifferentialDiffusion","_meta":{"title":"差异扩散DifferentialDiffusion"}},"54":{"inputs":{"images":["37",0]},"class_type":"PreviewImage","_meta":{"title":"预览图像"}},"83":{"inputs":{"style_model_name":"flux1-redux-dev.safetensors"},"class_type":"StyleModelLoader","_meta":{"title":"加载风格模型"}},"84":{"inputs":{"clip_name":"sigclip_vision_patch14_384.safetensors"},"class_type":"CLIPVisionLoader","_meta":{"title":"加载CLIP视觉"}},"85":{"inputs":{"image":"111.jpeg"},"class_type":"LoadImage","_meta":{"title":"加载图像"}},"86":{"inputs":{"downsampling_factor":1,"downsampling_function":"area","mode":"center crop (square)","weight":1,"autocrop_margin":0.1,"conditioning":["45",0],"style_model":["83",0],"clip_vision":["84",0],"image":["85",0]},"class_type":"ReduxAdvanced","_meta":{"title":"ReduxAdvanced"}},"88":{"inputs":{"direction":"right","match_image_size":true,"image1":["85",0],"image2":["10",0]},"class_type":"ImageConcanate","_meta":{"title":"Image Concatenate"}},"90":{"inputs":{"panel_width":["107",1],"panel_height":["107",0],"fill_color":"custom","fill_color_hex":"#000000"},"class_type":"CR Color Panel","_meta":{"title":"🌁 CR Color Panel"}},"92":{"inputs":{"direction":"right","match_image_size":true,"image1":["90",0],"image2":["93",0]},"class_type":"ImageConcanate","_meta":{"title":"Image Concatenate"}},"93":{"inputs":{"mask":["10",1]},"class_type":"MaskToImage","_meta":{"title":"遮罩转换为图像"}},"95":{"inputs":{"method":"intensity","image":["97",0]},"class_type":"Image To Mask","_meta":{"title":"Image To Mask"}},"96":{"inputs":{"size":1545,"interpolation_mode":"bicubic","image":["88",0]},"class_type":"JWImageResizeByLongerSide","_meta":{"title":"Image Resize by Longer Side"}},"97":{"inputs":{"size":1536,"interpolation_mode":"bicubic","image":["92",0]},"class_type":"JWImageResizeByLongerSide","_meta":{"title":"Image Resize by Longer Side"}},"98":{"inputs":{"images":["96",0]},"class_type":"PreviewImage","_meta":{"title":"预览图像"}},"99":{"inputs":{"images":["97",0]},"class_type":"PreviewImage","_meta":{"title":"预览图像"}},"103":{"inputs":{"vae_name":"ae.sft"},"class_type":"VAELoader","_meta":{"title":"加载VAE"}},"107":{"inputs":{"image":["85",0]},"class_type":"GetImageSize","_meta":{"title":"GetImageSize"}}},"extra_data":{"extra_pnginfo":{"workflow":{"id":"a5a0e040-be8e-4d31-991d-1740a8b1de9f","revision":0,"last_node_id":107,"last_link_id":147,"nodes":[{"id":97,"type":"JWImageResizeByLongerSide","pos":[-2130,8710],"size":[340.20001220703125,82],"flags":{},"order":24,"mode":0,"inputs":[{"label":"image","name":"image","type":"IMAGE","link":135}],"outputs":[{"label":"IMAGE","name":"IMAGE","type":"IMAGE","slot_index":0,"links":[136,138]}],"properties":{"Node name for S&R":"JWImageResizeByLongerSide","aux_id":"jamesWalker55/comfyui-various","ver":"5bd85aaf7616878471469c4ec7e11bbd0cef3bf2"},"widgets_values":[1536,"bicubic"]},{"id":101,"type":"Reroute","pos":[-2580,8750],"size":[75,26],"flags":{},"order":21,"mode":0,"inputs":[{"label":"","name":"","type":"*","link":139}],"outputs":[{"label":"","name":"","type":"IMAGE","slot_index":0,"links":[140]}],"properties":{"showOutputText":false,"horizontal":false}},{"id":93,"type":"MaskToImage","pos":[-2760,8790],"size":[210,26],"flags":{"collapsed":true},"order":17,"mode":0,"inputs":[{"label":"mask","name":"mask","type":"MASK","link":125}],"outputs":[{"label":"IMAGE","name":"IMAGE","type":"IMAGE","slot_index":0,"links":[139]}],"properties":{"Node name for S&R":"MaskToImage","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":[]},{"id":90,"type":"CR Color Panel","pos":[-2799.229248046875,8711.1064453125],"size":[315,150],"flags":{},"order":19,"mode":0,"inputs":[{"label":"panel_width","name":"panel_width","type":"INT","widget":{"name":"panel_width"},"link":146},{"label":"panel_height","name":"panel_height","type":"INT","widget":{"name":"panel_height"},"link":147}],"outputs":[{"label":"image","name":"image","type":"IMAGE","slot_index":0,"links":[124]},{"label":"show_help","name":"show_help","type":"STRING","links":null}],"properties":{"Node name for S&R":"CR Color Panel","aux_id":"Suzie1/ComfyUI_Comfyroll_CustomNodes","ver":"d78b780ae43fcf8c6b7c6505e6ffb4584281ceca"},"widgets_values":[512,512,"custom","#000000"]},{"id":88,"type":"ImageConcanate","pos":[-2790,8910],"size":[315,102],"flags":{"collapsed":true},"order":16,"mode":0,"inputs":[{"label":"image1","name":"image1","type":"IMAGE","link":118},{"label":"image2","name":"image2","type":"IMAGE","link":119}],"outputs":[{"label":"IMAGE","name":"IMAGE","type":"IMAGE","slot_index":0,"links":[132]}],"properties":{"Node name for S&R":"ImageConcanate","cnr_id":"comfyui-kjnodes","ver":"d126b62cebee81ea14ec06ea7cd7526999cb0554"},"widgets_values":["right",true]},{"id":96,"type":"JWImageResizeByLongerSide","pos":[-2380,8920],"size":[340.20001220703125,82],"flags":{"collapsed":true},"order":20,"mode":0,"inputs":[{"label":"image","name":"image","type":"IMAGE","link":132}],"outputs":[{"label":"IMAGE","name":"IMAGE","type":"IMAGE","slot_index":0,"links":[134,137]}],"properties":{"Node name for S&R":"JWImageResizeByLongerSide","aux_id":"jamesWalker55/comfyui-various","ver":"5bd85aaf7616878471469c4ec7e11bbd0cef3bf2"},"widgets_values":[1545,"bicubic"]},{"id":46,"type":"DifferentialDiffusion","pos":[-2570,7970],"size":[277.20001220703125,26],"flags":{"collapsed":true},"order":13,"mode":0,"inputs":[{"label":"model","name":"model","type":"MODEL","link":65}],"outputs":[{"label":"MODEL","name":"MODEL","type":"MODEL","slot_index":0,"links":[50]}],"properties":{"Node name for S&R":"DifferentialDiffusion","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":[]},{"id":86,"type":"ReduxAdvanced","pos":[-2100,7830],"size":[317.4000244140625,234],"flags":{"collapsed":true},"order":28,"mode":0,"inputs":[{"label":"conditioning","name":"conditioning","type":"CONDITIONING","link":111},{"label":"style_model","name":"style_model","type":"STYLE_MODEL","link":112},{"label":"clip_vision","name":"clip_vision","type":"CLIP_VISION","link":113},{"label":"image","name":"image","type":"IMAGE","link":143},{"label":"mask","name":"mask","shape":7,"type":"MASK","link":null}],"outputs":[{"label":"CONDITIONING","name":"CONDITIONING","type":"CONDITIONING","slot_index":0,"links":[114]},{"label":"IMAGE","name":"IMAGE","type":"IMAGE","slot_index":1,"links":[]},{"label":"MASK","name":"MASK","type":"MASK","links":null}],"properties":{"Node name for S&R":"ReduxAdvanced","aux_id":"kaibioinfo/ComfyUI_AdvancedRefluxControl","ver":"0a87efa252ae5e8f4af1225b0e19c867f908376a"},"widgets_values":[1,"area","center crop (square)",1,0.1]},{"id":41,"type":"FluxGuidance","pos":[-2200,8100],"size":[250,60],"flags":{"collapsed":true},"order":18,"mode":0,"inputs":[{"label":"conditioning","name":"conditioning","type":"CONDITIONING","link":59}],"outputs":[{"label":"CONDITIONING","name":"CONDITIONING","shape":3,"type":"CONDITIONING","slot_index":0,"links":[60]}],"properties":{"Node name for S&R":"FluxGuidance","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":[30]},{"id":45,"type":"InpaintModelConditioning","pos":[-2060,8100],"size":[302.4000244140625,138],"flags":{"collapsed":true},"order":27,"mode":0,"inputs":[{"label":"positive","name":"positive","type":"CONDITIONING","link":60},{"label":"negative","name":"negative","type":"CONDITIONING","link":61},{"label":"vae","name":"vae","type":"VAE","link":144},{"label":"pixels","name":"pixels","type":"IMAGE","link":134},{"label":"mask","name":"mask","type":"MASK","link":129}],"outputs":[{"label":"positive","name":"positive","type":"CONDITIONING","slot_index":0,"links":[111]},{"label":"negative","name":"negative","type":"CONDITIONING","slot_index":1,"links":[52]},{"label":"latent","name":"latent","type":"LATENT","slot_index":2,"links":[53]}],"properties":{"Node name for S&R":"InpaintModelConditioning","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":[false]},{"id":43,"type":"VAELoader","pos":[-1890,7800],"size":[315,58],"flags":{"collapsed":false},"order":0,"mode":0,"inputs":[],"outputs":[{"label":"VAE","name":"VAE","type":"VAE","slot_index":0,"links":[56]}],"properties":{"Node name for S&R":"VAELoader","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":["ae.sft"]},{"id":37,"type":"VAEDecode","pos":[-1690,8040],"size":[210,46],"flags":{"collapsed":true},"order":30,"mode":0,"inputs":[{"label":"samples","name":"samples","type":"LATENT","link":55},{"label":"vae","name":"vae","type":"VAE","link":56}],"outputs":[{"label":"IMAGE","name":"IMAGE","type":"IMAGE","slot_index":0,"links":[76]}],"properties":{"Node name for S&R":"VAEDecode","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":[]},{"id":35,"type":"KSampler","pos":[-1890,8010],"size":[424.28216552734375,494.49041748046875],"flags":{},"order":29,"mode":0,"inputs":[{"label":"model","name":"model","type":"MODEL","link":50},{"label":"positive","name":"positive","type":"CONDITIONING","link":114},{"label":"negative","name":"negative","type":"CONDITIONING","link":52},{"label":"latent_image","name":"latent_image","type":"LATENT","link":53}],"outputs":[{"label":"LATENT","name":"LATENT","type":"LATENT","slot_index":0,"links":[55]}],"properties":{"Node name for S&R":"KSampler","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":[228408831139036,"randomize",30,1,"euler","normal",1]},{"id":102,"type":"Reroute","pos":[-2270,8710],"size":[75,26],"flags":{},"order":14,"mode":0,"inputs":[{"label":"","name":"","type":"*","link":142}],"outputs":[{"label":"","name":"","type":"IMAGE","slot_index":0,"links":[143]}],"properties":{"showOutputText":false,"horizontal":false}},{"id":92,"type":"ImageConcanate","pos":[-2460,8710],"size":[315,102],"flags":{},"order":23,"mode":0,"inputs":[{"label":"image1","name":"image1","type":"IMAGE","link":124},{"label":"image2","name":"image2","type":"IMAGE","link":140}],"outputs":[{"label":"IMAGE","name":"IMAGE","type":"IMAGE","slot_index":0,"links":[135]}],"properties":{"Node name for S&R":"ImageConcanate","cnr_id":"comfyui-kjnodes","ver":"d126b62cebee81ea14ec06ea7cd7526999cb0554"},"widgets_values":["right",true]},{"id":103,"type":"VAELoader","pos":[-2190,8150],"size":[210,58],"flags":{"collapsed":false},"order":1,"mode":0,"inputs":[],"outputs":[{"label":"VAE","name":"VAE","type":"VAE","links":[144]}],"properties":{"Node name for S&R":"VAELoader","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":["ae.sft"]},{"id":36,"type":"CLIPTextEncode","pos":[-2460,8100],"size":[425.27801513671875,180.6060791015625],"flags":{"collapsed":true},"order":11,"mode":0,"inputs":[{"label":"clip","name":"clip","type":"CLIP","link":54}],"outputs":[{"label":"CONDITIONING","name":"CONDITIONING","type":"CONDITIONING","slot_index":0,"links":[61]}],"properties":{"Node name for S&R":"CLIPTextEncode","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":[""],"color":"#322","bgcolor":"#533"},{"id":105,"type":"Note","pos":[-2610,8390],"size":[390,110],"flags":{},"order":2,"mode":0,"inputs":[],"outputs":[],"properties":{},"widgets_values":["上面文本框可以写一下具体描述：\n比如：男子戴着眼镜，后者女子戴着眼镜等"],"color":"#432","bgcolor":"#653"},{"id":104,"type":"Note","pos":[-4423.1962890625,9810.134765625],"size":[480,88],"flags":{},"order":3,"mode":0,"inputs":[],"outputs":[],"properties":{},"widgets_values":["左侧产品图：白底，然后尺寸尽量和右侧模特图（目标图）保持一致"],"color":"#432","bgcolor":"#653"},{"id":106,"type":"Note","pos":[-3893.195556640625,9810.134765625],"size":[560,100],"flags":{},"order":4,"mode":0,"inputs":[],"outputs":[],"properties":{},"widgets_values":["需要操作：\n1.上传图像后，右键→在遮罩编辑器中打开→用画笔涂抹原有眼镜（物品），之后保存即可！"],"color":"#432","bgcolor":"#653"},{"id":107,"type":"GetImageSize","pos":[-3039,8755],"size":[210,46],"flags":{},"order":15,"mode":0,"inputs":[{"label":"image","name":"image","type":"IMAGE","link":145}],"outputs":[{"label":"width","name":"width","type":"INT","slot_index":0,"links":[147]},{"label":"height","name":"height","type":"INT","slot_index":1,"links":[146]}],"properties":{"Node name for S&R":"GetImageSize"},"widgets_values":[]},{"id":44,"type":"DualCLIPLoader","pos":[-2980,8070],"size":[345.1612243652344,122],"flags":{},"order":5,"mode":0,"inputs":[],"outputs":[{"label":"CLIP","name":"CLIP","type":"CLIP","links":[54,58]}],"properties":{"Node name for S&R":"DualCLIPLoader","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":["clip_l.safetensors","t5xxl_fp8_e4m3fn.safetensors","flux","default"]},{"id":84,"type":"CLIPVisionLoader","pos":[-3000,7800],"size":[370,60],"flags":{},"order":6,"mode":0,"inputs":[],"outputs":[{"label":"CLIP_VISION","name":"CLIP_VISION","type":"CLIP_VISION","slot_index":0,"links":[113]}],"properties":{"Node name for S&R":"CLIPVisionLoader","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":["sigclip_vision_patch14_384.safetensors"]},{"id":83,"type":"StyleModelLoader","pos":[-2600,7800],"size":[340,60],"flags":{},"order":7,"mode":0,"inputs":[],"outputs":[{"label":"STYLE_MODEL","name":"STYLE_MODEL","type":"STYLE_MODEL","links":[112]}],"properties":{"Node name for S&R":"StyleModelLoader","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":["flux1-redux-dev.safetensors"]},{"id":42,"type":"UNETLoader","pos":[-2990,7940],"size":[360,82],"flags":{},"order":8,"mode":0,"inputs":[],"outputs":[{"label":"MODEL","name":"MODEL","type":"MODEL","slot_index":0,"links":[65]}],"properties":{"Node name for S&R":"UNETLoader","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":["F.1-Fill-fp16_Inpaint&Outpaint_1.0.safetensors","fp8_e4m3fn_fast"]},{"id":99,"type":"PreviewImage","pos":[-1660,9020],"size":[420,750],"flags":{},"order":26,"mode":0,"inputs":[{"label":"images","name":"images","type":"IMAGE","link":138}],"outputs":[],"properties":{"Node name for S&R":"PreviewImage","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":[""]},{"id":95,"type":"Image To Mask","pos":[-1760,8710],"size":[315,58],"flags":{},"order":25,"mode":0,"inputs":[{"label":"image","name":"image","type":"IMAGE","link":136}],"outputs":[{"label":"MASK","name":"MASK","type":"MASK","slot_index":0,"links":[129]}],"properties":{"Node name for S&R":"Image To Mask","aux_id":"BadCafeCode/masquerade-nodes-comfyui","ver":"432cb4d146a391b387a0cd25ace824328b5b61cf"},"widgets_values":["intensity"]},{"id":98,"type":"PreviewImage","pos":[-2338.998046875,9248.4296875],"size":[450,750],"flags":{},"order":22,"mode":0,"inputs":[{"label":"images","name":"images","type":"IMAGE","link":137}],"outputs":[],"properties":{"Node name for S&R":"PreviewImage","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":[""]},{"id":54,"type":"PreviewImage","pos":[-4250,7720],"size":[1130,870],"flags":{},"order":31,"mode":0,"inputs":[{"label":"images","name":"images","type":"IMAGE","link":76}],"outputs":[],"properties":{"Node name for S&R":"PreviewImage","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":[""]},{"id":40,"type":"CLIPTextEncode","pos":[-2610,8070],"size":[390,280],"flags":{},"order":12,"mode":0,"inputs":[{"label":"clip","name":"clip","type":"CLIP","link":58}],"outputs":[{"label":"CONDITIONING","name":"CONDITIONING","type":"CONDITIONING","slot_index":0,"links":[59]}],"properties":{"Node name for S&R":"CLIPTextEncode","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":["The man is wearing glasses"],"color":"#232","bgcolor":"#353"},{"id":85,"type":"LoadImage","pos":[-4443.1962890625,8810.134765625],"size":[520,950],"flags":{},"order":9,"mode":0,"inputs":[],"outputs":[{"label":"IMAGE","name":"IMAGE","type":"IMAGE","slot_index":0,"links":[118,142,145]},{"label":"MASK","name":"MASK","type":"MASK","slot_index":1,"links":[]}],"properties":{"Node name for S&R":"LoadImage","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":["111.jpeg","image",""]},{"id":10,"type":"LoadImage","pos":[-3888.195556640625,8817.134765625],"size":[550,950],"flags":{},"order":10,"mode":0,"inputs":[],"outputs":[{"label":"IMAGE","name":"IMAGE","type":"IMAGE","slot_index":0,"links":[119]},{"label":"MASK","name":"MASK","type":"MASK","slot_index":1,"links":[125]}],"properties":{"Node name for S&R":"LoadImage","cnr_id":"comfy-core","ver":"0.3.24"},"widgets_values":["clipspace/clipspace-mask-1961925.8999999762.png [input]","image",""],"color":"#232","bgcolor":"#353"}],"links":[[50,46,0,35,0,"MODEL"],[52,45,1,35,2,"CONDITIONING"],[53,45,2,35,3,"LATENT"],[54,44,0,36,0,"CLIP"],[55,35,0,37,0,"LATENT"],[56,43,0,37,1,"VAE"],[58,44,0,40,0,"CLIP"],[59,40,0,41,0,"CONDITIONING"],[60,41,0,45,0,"CONDITIONING"],[61,36,0,45,1,"CONDITIONING"],[65,42,0,46,0,"MODEL"],[76,37,0,54,0,"IMAGE"],[111,45,0,86,0,"CONDITIONING"],[112,83,0,86,1,"STYLE_MODEL"],[113,84,0,86,2,"CLIP_VISION"],[114,86,0,35,1,"CONDITIONING"],[118,85,0,88,0,"IMAGE"],[119,10,0,88,1,"IMAGE"],[124,90,0,92,0,"IMAGE"],[125,10,1,93,0,"MASK"],[129,95,0,45,4,"MASK"],[132,88,0,96,0,"IMAGE"],[134,96,0,45,3,"IMAGE"],[135,92,0,97,0,"IMAGE"],[136,97,0,95,0,"IMAGE"],[137,96,0,98,0,"IMAGE"],[138,97,0,99,0,"IMAGE"],[139,93,0,101,0,"*"],[140,101,0,92,1,"IMAGE"],[142,85,0,102,0,"*"],[143,102,0,86,3,"IMAGE"],[144,103,0,45,2,"VAE"],[145,85,0,107,0,"IMAGE"],[146,107,1,90,0,"INT"],[147,107,0,90,1,"INT"]],"groups":[{"id":1,"title":"标准Redux+fill inpaint","bounding":[-3070,7590,1850,1030],"color":"#3f789e","font_size":22,"flags":{}},{"id":2,"title":"左侧图像：白底产品图","bounding":[-4473.1962890625,8660.134765625,1200,1280],"color":"#8A8","font_size":50,"flags":{}},{"id":3,"title":"图像拼接","bounding":[-3070,8630,1860,1290],"color":"#8AA","font_size":24,"flags":{}},{"id":4,"title":"输出结果图","bounding":[-4280,7590,1190,1030],"color":"#b58b2a","font_size":50,"flags":{}}],"config":{},"extra":{"ds":{"scale":0.29408349370551534,"offset":[5910.519255484234,-7434.752229016367]},"ue_links":[],"workspace_info":{"id":"teQ3_Nr1kf7g2e0Tnjehx"},"0246.VERSION":[0,0,4],"VHS_latentpreview":false,"VHS_latentpreviewrate":0},"version":0.4,"widget_idx_map":{"35":{"seed":0,"sampler_name":4,"scheduler":5}}}}}}'