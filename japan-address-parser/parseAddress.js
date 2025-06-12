
const japa = require('jp-address-parser');


(async () => {
  try {
    const address = process.argv[2];
    var result = { success: true }
    result = { ...result, ...await japa.parse(address) }
    console.log(JSON.stringify(result))
  }
  catch {
    console.log(JSON.stringify({ success: false }))
  }
  /*
  { prefecture: '東京都',
    city: '北区',
    town: '東十条',
    chome: 6,
    ban: 28,
    go: 70,
    left: '' }
  */
})()


