# Lychee JavaScript SDK

The official JavaScript/TypeScript client library for Lychee — the universal local inference engine.

## Installation

```bash
npm install @lychee/lychee
```

## Quick Start

```javascript
import lychee from '@lychee/lychee'

// Simple chat completion
const response = await lychee.chat({
  model: 'gemma3',
  messages: [{ role: 'user', content: 'Explain quantum physics in one sentence.' }]
})
console.log(response.message.content)
```

## Custom Client Configuration

```javascript
import { Lychee } from '@lychee/lychee'

const client = new Lychee({ host: 'http://localhost:11434' })

// Generate text
const response = await client.generate({
  model: 'gemma3',
  prompt: 'Write a poem about trees.'
})
console.log(response.response)

// List local models
const models = await client.list()
for (const model of models.models) {
  console.log(model.name)
}
```
