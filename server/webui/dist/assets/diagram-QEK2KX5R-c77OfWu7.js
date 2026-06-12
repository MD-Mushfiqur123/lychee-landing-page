var I=Object.defineProperty;var b=Object.getOwnPropertySymbols;var _=Object.prototype.hasOwnProperty,E=Object.prototype.propertyIsEnumerable;var M=(a,t,e)=>t in a?I(a,t,{enumerable:!0,configurable:!0,writable:!0,value:e}):a[t]=e,C=(a,t)=>{for(var e in t||(t={}))_.call(t,e)&&M(a,e,t[e]);if(b)for(var e of b(t))E.call(t,e)&&M(a,e,t[e]);return a};import{_ as l,s as D,g as G,t as z,q as P,a as H,b as V,K as W,z as B,F as w,G as A,H as j,l as q,a1 as K}from"./mermaid.core-BPnNUrQO.js";import{p as N}from"./chunk-4BX2VUAB-Bo2iW_sx.js";import{p as U}from"./treemap-KMMF4GRG-C372xg7q.js";import"./index-E9nPRTPp.js";import"./_baseUniq-El2NPsBe.js";import"./_basePickBy-COMTkAx8.js";import"./clone-CRI1sbiQ.js";var x={showLegend:!0,ticks:5,max:null,min:0,graticule:"circle"},L={axes:[],curves:[],options:x},h=structuredClone(L),X=j.radar,Y=l(()=>w(C(C({},X),A().radar)),"getConfig"),T=l(()=>h.axes,"getAxes"),Z=l(()=>h.curves,"getCurves"),J=l(()=>h.options,"getOptions"),Q=l(a=>{h.axes=a.map(t=>{var e;return{name:t.name,label:(e=t.label)!=null?e:t.name}})},"setAxes"),tt=l(a=>{h.curves=a.map(t=>{var e;return{name:t.name,label:(e=t.label)!=null?e:t.name,entries:et(t.entries)}})},"setCurves"),et=l(a=>{if(a[0].axis==null)return a.map(e=>e.value);const t=T();if(t.length===0)throw new Error("Axes must be populated before curves for reference entries");return t.map(e=>{const r=a.find(s=>{var o;return((o=s.axis)==null?void 0:o.$refText)===e.name});if(r===void 0)throw new Error("Missing entry for axis "+e.label);return r.value})},"computeCurveEntries"),at=l(a=>{var e,r,s,o,i,n,c,d,u,p;const t=a.reduce((g,m)=>(g[m.name]=m,g),{});h.options={showLegend:(r=(e=t.showLegend)==null?void 0:e.value)!=null?r:x.showLegend,ticks:(o=(s=t.ticks)==null?void 0:s.value)!=null?o:x.ticks,max:(n=(i=t.max)==null?void 0:i.value)!=null?n:x.max,min:(d=(c=t.min)==null?void 0:c.value)!=null?d:x.min,graticule:(p=(u=t.graticule)==null?void 0:u.value)!=null?p:x.graticule}},"setOptions"),rt=l(()=>{B(),h=structuredClone(L)},"clear"),f={getAxes:T,getCurves:Z,getOptions:J,setAxes:Q,setCurves:tt,setOptions:at,getConfig:Y,clear:rt,setAccTitle:V,getAccTitle:H,setDiagramTitle:P,getDiagramTitle:z,getAccDescription:G,setAccDescription:D},st=l(a=>{N(a,f);const{axes:t,curves:e,options:r}=a;f.setAxes(t),f.setCurves(e),f.setOptions(r)},"populate"),nt={parse:l(async a=>{const t=await U("radar",a);q.debug(t),st(t)},"parse")},ot=l((a,t,e,r)=>{var $;const s=r.db,o=s.getAxes(),i=s.getCurves(),n=s.getOptions(),c=s.getConfig(),d=s.getDiagramTitle(),u=W(t),p=it(u,c),g=($=n.max)!=null?$:Math.max(...i.map(y=>Math.max(...y.entries))),m=n.min,v=Math.min(c.width,c.height)/2;lt(p,o,v,n.ticks,n.graticule),ct(p,o,v,c),O(p,o,i,m,g,n.graticule,c),R(p,i,n.showLegend,c),p.append("text").attr("class","radarTitle").text(d).attr("x",0).attr("y",-c.height/2-c.marginTop)},"draw"),it=l((a,t)=>{const e=t.width+t.marginLeft+t.marginRight,r=t.height+t.marginTop+t.marginBottom,s={x:t.marginLeft+t.width/2,y:t.marginTop+t.height/2};return a.attr("viewbox",`0 0 ${e} ${r}`).attr("width",e).attr("height",r),a.append("g").attr("transform",`translate(${s.x}, ${s.y})`)},"drawFrame"),lt=l((a,t,e,r,s)=>{if(s==="circle")for(let o=0;o<r;o++){const i=e*(o+1)/r;a.append("circle").attr("r",i).attr("class","radarGraticule")}else if(s==="polygon"){const o=t.length;for(let i=0;i<r;i++){const n=e*(i+1)/r,c=t.map((d,u)=>{const p=2*u*Math.PI/o-Math.PI/2,g=n*Math.cos(p),m=n*Math.sin(p);return`${g},${m}`}).join(" ");a.append("polygon").attr("points",c).attr("class","radarGraticule")}}},"drawGraticule"),ct=l((a,t,e,r)=>{const s=t.length;for(let o=0;o<s;o++){const i=t[o].label,n=2*o*Math.PI/s-Math.PI/2;a.append("line").attr("x1",0).attr("y1",0).attr("x2",e*r.axisScaleFactor*Math.cos(n)).attr("y2",e*r.axisScaleFactor*Math.sin(n)).attr("class","radarAxisLine"),a.append("text").text(i).attr("x",e*r.axisLabelFactor*Math.cos(n)).attr("y",e*r.axisLabelFactor*Math.sin(n)).attr("class","radarAxisLabel")}},"drawAxes");function O(a,t,e,r,s,o,i){const n=t.length,c=Math.min(i.width,i.height)/2;e.forEach((d,u)=>{if(d.entries.length!==n)return;const p=d.entries.map((g,m)=>{const v=2*Math.PI*m/n-Math.PI/2,$=S(g,r,s,c),y=$*Math.cos(v),F=$*Math.sin(v);return{x:y,y:F}});o==="circle"?a.append("path").attr("d",k(p,i.curveTension)).attr("class",`radarCurve-${u}`):o==="polygon"&&a.append("polygon").attr("points",p.map(g=>`${g.x},${g.y}`).join(" ")).attr("class",`radarCurve-${u}`)})}l(O,"drawCurves");function S(a,t,e,r){const s=Math.min(Math.max(a,t),e);return r*(s-t)/(e-t)}l(S,"relativeRadius");function k(a,t){const e=a.length;let r=`M${a[0].x},${a[0].y}`;for(let s=0;s<e;s++){const o=a[(s-1+e)%e],i=a[s],n=a[(s+1)%e],c=a[(s+2)%e],d={x:i.x+(n.x-o.x)*t,y:i.y+(n.y-o.y)*t},u={x:n.x-(c.x-i.x)*t,y:n.y-(c.y-i.y)*t};r+=` C${d.x},${d.y} ${u.x},${u.y} ${n.x},${n.y}`}return`${r} Z`}l(k,"closedRoundCurve");function R(a,t,e,r){if(!e)return;const s=(r.width/2+r.marginRight)*3/4,o=-(r.height/2+r.marginTop)*3/4,i=20;t.forEach((n,c)=>{const d=a.append("g").attr("transform",`translate(${s}, ${o+c*i})`);d.append("rect").attr("width",12).attr("height",12).attr("class",`radarLegendBox-${c}`),d.append("text").attr("x",16).attr("y",0).attr("class","radarLegendText").text(n.label)})}l(R,"drawLegend");var dt={draw:ot},pt=l((a,t)=>{let e="";for(let r=0;r<a.THEME_COLOR_LIMIT;r++){const s=a[`cScale${r}`];e+=`
		.radarCurve-${r} {
			color: ${s};
			fill: ${s};
			fill-opacity: ${t.curveOpacity};
			stroke: ${s};
			stroke-width: ${t.curveStrokeWidth};
		}
		.radarLegendBox-${r} {
			fill: ${s};
			fill-opacity: ${t.curveOpacity};
			stroke: ${s};
		}
		`}return e},"genIndexStyles"),ut=l(a=>{const t=K(),e=A(),r=w(t,e.themeVariables),s=w(r.radar,a);return{themeVariables:r,radarOptions:s}},"buildRadarStyleOptions"),gt=l(({radar:a}={})=>{const{themeVariables:t,radarOptions:e}=ut(a);return`
	.radarTitle {
		font-size: ${t.fontSize};
		color: ${t.titleColor};
		dominant-baseline: hanging;
		text-anchor: middle;
	}
	.radarAxisLine {
		stroke: ${e.axisColor};
		stroke-width: ${e.axisStrokeWidth};
	}
	.radarAxisLabel {
		dominant-baseline: middle;
		text-anchor: middle;
		font-size: ${e.axisLabelFontSize}px;
		color: ${e.axisColor};
	}
	.radarGraticule {
		fill: ${e.graticuleColor};
		fill-opacity: ${e.graticuleOpacity};
		stroke: ${e.graticuleColor};
		stroke-width: ${e.graticuleStrokeWidth};
	}
	.radarLegendText {
		text-anchor: start;
		font-size: ${e.legendFontSize}px;
		dominant-baseline: hanging;
	}
	${pt(t,e)}
	`},"styles"),wt={parser:nt,db:f,renderer:dt,styles:gt};export{wt as diagram};
