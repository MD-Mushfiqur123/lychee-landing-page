var D=Object.defineProperty;var w=Object.getOwnPropertySymbols;var T=Object.prototype.hasOwnProperty,z=Object.prototype.propertyIsEnumerable;var v=(e,t,a)=>t in e?D(e,t,{enumerable:!0,configurable:!0,writable:!0,value:a}):e[t]=a,m=(e,t)=>{for(var a in t||(t={}))T.call(t,a)&&v(e,a,t[a]);if(w)for(var a of w(t))z.call(t,a)&&v(e,a,t[a]);return e};import{_ as u,F as x,K as F,e as P,l as y,b as E,a as A,q as W,t as _,g as N,s as L,G as M,H as Y,z as G}from"./mermaid.core-BPnNUrQO.js";import{p as H}from"./chunk-4BX2VUAB-Bo2iW_sx.js";import{p as I}from"./treemap-KMMF4GRG-C372xg7q.js";import"./index-E9nPRTPp.js";import"./_baseUniq-El2NPsBe.js";import"./_basePickBy-COMTkAx8.js";import"./clone-CRI1sbiQ.js";var K=Y.packet,b,$=(b=class{constructor(){this.packet=[],this.setAccTitle=E,this.getAccTitle=A,this.setDiagramTitle=W,this.getDiagramTitle=_,this.getAccDescription=N,this.setAccDescription=L}getConfig(){const t=x(m(m({},K),M().packet));return t.showBits&&(t.paddingY+=10),t}getPacket(){return this.packet}pushWord(t){t.length>0&&this.packet.push(t)}clear(){G(),this.packet=[]}},u(b,"PacketDB"),b),O=1e4,j=u((e,t)=>{H(e,t);let a=-1,o=[],l=1;const{bitsPerRow:c}=t.getConfig();for(let{start:r,end:s,bits:n,label:d}of e.blocks){if(r!==void 0&&s!==void 0&&s<r)throw new Error(`Packet block ${r} - ${s} is invalid. End must be greater than start.`);if(r!=null||(r=a+1),r!==a+1)throw new Error(`Packet block ${r} - ${s!=null?s:r} is not contiguous. It should start from ${a+1}.`);if(n===0)throw new Error(`Packet block ${r} is invalid. Cannot have a zero bit field.`);for(s!=null||(s=r+(n!=null?n:1)-1),n!=null||(n=s-r+1),a=s,y.debug(`Packet block ${r} - ${a} with label ${d}`);o.length<=c+1&&t.getPacket().length<O;){const[p,i]=q({start:r,end:s,bits:n,label:d},l,c);if(o.push(p),p.end+1===l*c&&(t.pushWord(o),o=[],l++),!i)break;({start:r,end:s,bits:n,label:d}=i)}}t.pushWord(o)},"populate"),q=u((e,t,a)=>{if(e.start===void 0)throw new Error("start should have been set during first phase");if(e.end===void 0)throw new Error("end should have been set during first phase");if(e.start>e.end)throw new Error(`Block start ${e.start} is greater than block end ${e.end}.`);if(e.end+1<=t*a)return[e,void 0];const o=t*a-1,l=t*a;return[{start:e.start,end:o,label:e.label,bits:o-e.start},{start:l,end:e.end,label:e.label,bits:e.end-l}]},"getNextFittingBlock"),B={parser:{yy:void 0},parse:u(async e=>{var o;const t=await I("packet",e),a=(o=B.parser)==null?void 0:o.yy;if(!(a instanceof $))throw new Error("parser.parser?.yy was not a PacketDB. This is due to a bug within Mermaid, please report this issue at https://github.com/mermaid-js/mermaid/issues.");y.debug(t),j(t,a)},"parse")},R=u((e,t,a,o)=>{const l=o.db,c=l.getConfig(),{rowHeight:r,paddingY:s,bitWidth:n,bitsPerRow:d}=c,p=l.getPacket(),i=l.getDiagramTitle(),h=r+s,g=h*(p.length+1)-(i?0:r),k=n*d+2,f=F(t);f.attr("viewbox",`0 0 ${k} ${g}`),P(f,g,k,c.useMaxWidth);for(const[C,S]of p.entries())U(f,S,C,c);f.append("text").text(i).attr("x",k/2).attr("y",g-h/2).attr("dominant-baseline","middle").attr("text-anchor","middle").attr("class","packetTitle")},"draw"),U=u((e,t,a,{rowHeight:o,paddingX:l,paddingY:c,bitWidth:r,bitsPerRow:s,showBits:n})=>{const d=e.append("g"),p=a*(o+c)+c;for(const i of t){const h=i.start%s*r+1,g=(i.end-i.start+1)*r-l;if(d.append("rect").attr("x",h).attr("y",p).attr("width",g).attr("height",o).attr("class","packetBlock"),d.append("text").attr("x",h+g/2).attr("y",p+o/2).attr("class","packetLabel").attr("dominant-baseline","middle").attr("text-anchor","middle").text(i.label),!n)continue;const k=i.end===i.start,f=p-2;d.append("text").attr("x",h+(k?g/2:0)).attr("y",f).attr("class","packetByte start").attr("dominant-baseline","auto").attr("text-anchor",k?"middle":"start").text(i.start),k||d.append("text").attr("x",h+g).attr("y",f).attr("class","packetByte end").attr("dominant-baseline","auto").attr("text-anchor","end").text(i.end)}},"drawWord"),X={draw:R},J={byteFontSize:"10px",startByteColor:"black",endByteColor:"black",labelColor:"black",labelFontSize:"12px",titleColor:"black",titleFontSize:"14px",blockStrokeColor:"black",blockStrokeWidth:"1",blockFillColor:"#efefef"},Q=u(({packet:e}={})=>{const t=x(J,e);return`
	.packetByte {
		font-size: ${t.byteFontSize};
	}
	.packetByte.start {
		fill: ${t.startByteColor};
	}
	.packetByte.end {
		fill: ${t.endByteColor};
	}
	.packetLabel {
		fill: ${t.labelColor};
		font-size: ${t.labelFontSize};
	}
	.packetTitle {
		fill: ${t.titleColor};
		font-size: ${t.titleFontSize};
	}
	.packetBlock {
		stroke: ${t.blockStrokeColor};
		stroke-width: ${t.blockStrokeWidth};
		fill: ${t.blockFillColor};
	}
	`},"styles"),it={parser:B,get db(){return new $},renderer:X,styles:Q};export{it as diagram};
