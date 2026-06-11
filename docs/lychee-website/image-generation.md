# Image Generation

Lychee integrates diffusion model pipelines to allow local image generation alongside text generation.

## Supported Models

- `sdxl` (Stable Diffusion XL)
- `sd15` (Stable Diffusion v1.5)
- Custom ONNX and CoreML exported diffusion models.

## Usage

```bash
lychee generate sdxl "A highly detailed portrait of a futuristic city, cyberpunk style" --output city.png
```

## Memory Requirements

For SDXL, we recommend at least 16GB of system RAM and 8GB of VRAM. Lychee will dynamically quantize weights if VRAM is insufficient.
