import{d as we,c as Ce,a as ye,h as x,b as w,e as i,f as I,g as _,i as Se,j as ze,k as Re,l as ee,r as _e,N as Me,o as De,m as Ee,u as Ie,n as y,p as Te,q as Be,s as oe,t as $e,v as Ne,w as J,x as Le,y as j,z as q,A as Ue,B as Ve,C as Pe,D as Fe,E as Ke,F as je,G as H,H as z,I as C,T as He,J as Oe,K as Ae,L as M,M as b,O as We,P as Q,Q as O,R as X,S as $,U as Ge,V as A,W as Z}from"./index-co_nq71q.js";import{u as Ye,N as Je,a as L,b as U,_ as qe}from"./_plugin-vue_export-helper-BATpDewr.js";const Qe={sizeSmall:"14px",sizeMedium:"16px",sizeLarge:"18px",labelPadding:"0 8px",labelFontWeight:"400"};function Xe(n){const{baseColor:s,inputColorDisabled:c,cardColor:g,modalColor:R,popoverColor:D,textColorDisabled:d,borderColor:p,primaryColor:l,textColor2:r,fontSizeSmall:h,fontSizeMedium:v,fontSizeLarge:u,borderRadiusSmall:m,lineHeight:f}=n;return Object.assign(Object.assign({},Qe),{labelLineHeight:f,fontSizeSmall:h,fontSizeMedium:v,fontSizeLarge:u,borderRadius:m,color:s,colorChecked:l,colorDisabled:c,colorDisabledChecked:c,colorTableHeader:g,colorTableHeaderModal:R,colorTableHeaderPopover:D,checkMarkColor:s,checkMarkColorDisabled:d,checkMarkColorDisabledChecked:d,border:`1px solid ${p}`,borderDisabled:`1px solid ${p}`,borderDisabledChecked:`1px solid ${p}`,borderChecked:`1px solid ${l}`,borderFocus:`1px solid ${l}`,boxShadowFocus:`0 0 0 2px ${Ce(l,{alpha:.3})}`,textColor:r,textColorDisabled:d})}const Ze={common:we,self:Xe},eo=ye("n-checkbox-group"),oo=()=>x("svg",{viewBox:"0 0 64 64",class:"check-icon"},x("path",{d:"M50.42,16.76L22.34,39.45l-8.1-11.46c-1.12-1.58-3.3-1.96-4.88-0.84c-1.58,1.12-1.95,3.3-0.84,4.88l10.26,14.51  c0.56,0.79,1.42,1.31,2.38,1.45c0.16,0.02,0.32,0.03,0.48,0.03c0.8,0,1.57-0.27,2.2-0.78l30.99-25.03c1.5-1.21,1.74-3.42,0.52-4.92  C54.13,15.78,51.93,15.55,50.42,16.76z"})),ao=()=>x("svg",{viewBox:"0 0 100 100",class:"line-icon"},x("path",{d:"M80.2,55.5H21.4c-2.8,0-5.1-2.5-5.1-5.5l0,0c0-3,2.3-5.5,5.1-5.5h58.7c2.8,0,5.1,2.5,5.1,5.5l0,0C85.2,53.1,82.9,55.5,80.2,55.5z"})),ro=w([i("checkbox",`
 font-size: var(--n-font-size);
 outline: none;
 cursor: pointer;
 display: inline-flex;
 flex-wrap: nowrap;
 align-items: flex-start;
 word-break: break-word;
 line-height: var(--n-size);
 --n-merged-color-table: var(--n-color-table);
 `,[I("show-label","line-height: var(--n-label-line-height);"),w("&:hover",[i("checkbox-box",[_("border","border: var(--n-border-checked);")])]),w("&:focus:not(:active)",[i("checkbox-box",[_("border",`
 border: var(--n-border-focus);
 box-shadow: var(--n-box-shadow-focus);
 `)])]),I("inside-table",[i("checkbox-box",`
 background-color: var(--n-merged-color-table);
 `)]),I("checked",[i("checkbox-box",`
 background-color: var(--n-color-checked);
 `,[i("checkbox-icon",[w(".check-icon",`
 opacity: 1;
 transform: scale(1);
 `)])])]),I("indeterminate",[i("checkbox-box",[i("checkbox-icon",[w(".check-icon",`
 opacity: 0;
 transform: scale(.5);
 `),w(".line-icon",`
 opacity: 1;
 transform: scale(1);
 `)])])]),I("checked, indeterminate",[w("&:focus:not(:active)",[i("checkbox-box",[_("border",`
 border: var(--n-border-checked);
 box-shadow: var(--n-box-shadow-focus);
 `)])]),i("checkbox-box",`
 background-color: var(--n-color-checked);
 border-left: 0;
 border-top: 0;
 `,[_("border",{border:"var(--n-border-checked)"})])]),I("disabled",{cursor:"not-allowed"},[I("checked",[i("checkbox-box",`
 background-color: var(--n-color-disabled-checked);
 `,[_("border",{border:"var(--n-border-disabled-checked)"}),i("checkbox-icon",[w(".check-icon, .line-icon",{fill:"var(--n-check-mark-color-disabled-checked)"})])])]),i("checkbox-box",`
 background-color: var(--n-color-disabled);
 `,[_("border",`
 border: var(--n-border-disabled);
 `),i("checkbox-icon",[w(".check-icon, .line-icon",`
 fill: var(--n-check-mark-color-disabled);
 `)])]),_("label",`
 color: var(--n-text-color-disabled);
 `)]),i("checkbox-box-wrapper",`
 position: relative;
 width: var(--n-size);
 flex-shrink: 0;
 flex-grow: 0;
 user-select: none;
 -webkit-user-select: none;
 `),i("checkbox-box",`
 position: absolute;
 left: 0;
 top: 50%;
 transform: translateY(-50%);
 height: var(--n-size);
 width: var(--n-size);
 display: inline-block;
 box-sizing: border-box;
 border-radius: var(--n-border-radius);
 background-color: var(--n-color);
 transition: background-color 0.3s var(--n-bezier);
 `,[_("border",`
 transition:
 border-color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 border-radius: inherit;
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 border: var(--n-border);
 `),i("checkbox-icon",`
 display: flex;
 align-items: center;
 justify-content: center;
 position: absolute;
 left: 1px;
 right: 1px;
 top: 1px;
 bottom: 1px;
 `,[w(".check-icon, .line-icon",`
 width: 100%;
 fill: var(--n-check-mark-color);
 opacity: 0;
 transform: scale(0.5);
 transform-origin: center;
 transition:
 fill 0.3s var(--n-bezier),
 transform 0.3s var(--n-bezier),
 opacity 0.3s var(--n-bezier),
 border-color 0.3s var(--n-bezier);
 `),Se({left:"1px",top:"1px"})])]),_("label",`
 color: var(--n-text-color);
 transition: color .3s var(--n-bezier);
 user-select: none;
 -webkit-user-select: none;
 padding: var(--n-label-padding);
 font-weight: var(--n-label-font-weight);
 `,[w("&:empty",{display:"none"})])]),ze(i("checkbox",`
 --n-merged-color-table: var(--n-color-table-modal);
 `)),Re(i("checkbox",`
 --n-merged-color-table: var(--n-color-table-popover);
 `))]),lo=Object.assign(Object.assign({},oe.props),{size:String,checked:{type:[Boolean,String,Number],default:void 0},defaultChecked:{type:[Boolean,String,Number],default:!1},value:[String,Number],disabled:{type:Boolean,default:void 0},indeterminate:Boolean,label:String,focusable:{type:Boolean,default:!0},checkedValue:{type:[Boolean,String,Number],default:!0},uncheckedValue:{type:[Boolean,String,Number],default:!1},"onUpdate:checked":[Function,Array],onUpdateChecked:[Function,Array],privateInsideTable:Boolean,onChange:[Function,Array]}),no=ee({name:"Checkbox",props:lo,setup(n){const s=Ee(eo,null),c=y(null),{mergedClsPrefixRef:g,inlineThemeDisabled:R,mergedRtlRef:D}=Ie(n),d=y(n.defaultChecked),p=Ue(n,"checked"),l=Ye(p,d),r=Te(()=>{if(s){const e=s.valueSetRef.value;return e&&n.value!==void 0?e.has(n.value):!1}else return l.value===n.checkedValue}),h=Be(n,{mergedSize(e){const{size:k}=n;if(k!==void 0)return k;if(s){const{value:a}=s.mergedSizeRef;if(a!==void 0)return a}if(e){const{mergedSize:a}=e;if(a!==void 0)return a.value}return"medium"},mergedDisabled(e){const{disabled:k}=n;if(k!==void 0)return k;if(s){if(s.disabledRef.value)return!0;const{maxRef:{value:a},checkedCountRef:o}=s;if(a!==void 0&&o.value>=a&&!r.value)return!0;const{minRef:{value:t}}=s;if(t!==void 0&&o.value<=t&&r.value)return!0}return e?e.disabled.value:!1}}),{mergedDisabledRef:v,mergedSizeRef:u}=h,m=oe("Checkbox","-checkbox",ro,Ze,n,g);function f(e){if(s&&n.value!==void 0)s.toggleCheckbox(!r.value,n.value);else{const{onChange:k,"onUpdate:checked":a,onUpdateChecked:o}=n,{nTriggerFormInput:t,nTriggerFormChange:K}=h,B=r.value?n.uncheckedValue:n.checkedValue;a&&j(a,B,e),o&&j(o,B,e),k&&j(k,B,e),t(),K(),d.value=B}}function S(e){v.value||f(e)}function T(e){if(!v.value)switch(e.key){case" ":case"Enter":f(e)}}function N(e){switch(e.key){case" ":e.preventDefault()}}const P={focus:()=>{var e;(e=c.value)===null||e===void 0||e.focus()},blur:()=>{var e;(e=c.value)===null||e===void 0||e.blur()}},F=$e("Checkbox",D,g),V=J(()=>{const{value:e}=u,{common:{cubicBezierEaseInOut:k},self:{borderRadius:a,color:o,colorChecked:t,colorDisabled:K,colorTableHeader:B,colorTableHeaderModal:ae,colorTableHeaderPopover:re,checkMarkColor:le,checkMarkColorDisabled:ne,border:te,borderFocus:ie,borderDisabled:se,borderChecked:ce,boxShadowFocus:de,textColor:ue,textColorDisabled:be,checkMarkColorDisabledChecked:he,colorDisabledChecked:ve,borderDisabledChecked:me,labelPadding:fe,labelLineHeight:ke,labelFontWeight:ge,[q("fontSize",e)]:pe,[q("size",e)]:xe}}=m.value;return{"--n-label-line-height":ke,"--n-label-font-weight":ge,"--n-size":xe,"--n-bezier":k,"--n-border-radius":a,"--n-border":te,"--n-border-checked":ce,"--n-border-focus":ie,"--n-border-disabled":se,"--n-border-disabled-checked":me,"--n-box-shadow-focus":de,"--n-color":o,"--n-color-checked":t,"--n-color-table":B,"--n-color-table-modal":ae,"--n-color-table-popover":re,"--n-color-disabled":K,"--n-color-disabled-checked":ve,"--n-text-color":ue,"--n-text-color-disabled":be,"--n-check-mark-color":le,"--n-check-mark-color-disabled":ne,"--n-check-mark-color-disabled-checked":he,"--n-font-size":pe,"--n-label-padding":fe}}),E=R?Ne("checkbox",J(()=>u.value[0]),V,n):void 0;return Object.assign(h,P,{rtlEnabled:F,selfRef:c,mergedClsPrefix:g,mergedDisabled:v,renderedChecked:r,mergedTheme:m,labelId:Le(),handleClick:S,handleKeyUp:T,handleKeyDown:N,cssVars:R?void 0:V,themeClass:E==null?void 0:E.themeClass,onRender:E==null?void 0:E.onRender})},render(){var n;const{$slots:s,renderedChecked:c,mergedDisabled:g,indeterminate:R,privateInsideTable:D,cssVars:d,labelId:p,label:l,mergedClsPrefix:r,focusable:h,handleKeyUp:v,handleKeyDown:u,handleClick:m}=this;(n=this.onRender)===null||n===void 0||n.call(this);const f=_e(s.default,S=>l||S?x("span",{class:`${r}-checkbox__label`,id:p},l||S):null);return x("div",{ref:"selfRef",class:[`${r}-checkbox`,this.themeClass,this.rtlEnabled&&`${r}-checkbox--rtl`,c&&`${r}-checkbox--checked`,g&&`${r}-checkbox--disabled`,R&&`${r}-checkbox--indeterminate`,D&&`${r}-checkbox--inside-table`,f&&`${r}-checkbox--show-label`],tabindex:g||!h?void 0:0,role:"checkbox","aria-checked":R?"mixed":c,"aria-labelledby":p,style:d,onKeyup:v,onKeydown:u,onClick:m,onMousedown:()=>{De("selectstart",window,S=>{S.preventDefault()},{once:!0})}},x("div",{class:`${r}-checkbox-box-wrapper`}," ",x("div",{class:`${r}-checkbox-box`},x(Me,null,{default:()=>this.indeterminate?x("div",{key:"indeterminate",class:`${r}-checkbox-icon`},ao()):x("div",{key:"check",class:`${r}-checkbox-icon`},oo())}),x("div",{class:`${r}-checkbox-box__border`}))),f)}}),to={class:"login-subtitle"},io={key:1,class:"login-options-placeholder"},W="ops_remember_login",G="ops_remember_username",Y="ops_remember_password",so=ee({__name:"LoginView",setup(n){const s=Ae(),c=Ve(),g=Pe();function R(){return localStorage.getItem(W)==="1"?{remembered:!0,username:localStorage.getItem(G)||"",password:localStorage.getItem(Y)||""}:{remembered:!1,username:"",password:""}}function D(a,o,t){if(!a){localStorage.removeItem(W),localStorage.removeItem(G),localStorage.removeItem(Y);return}localStorage.setItem(W,"1"),localStorage.setItem(G,o),localStorage.setItem(Y,t)}const d=R(),p={username:d.username,password:d.password,remember:d.remembered},l=y("login"),r=y(d.username),h=y(d.password),v=y(""),u=y(""),m=y(""),f=y(d.remembered),S=y(!1),T=y(typeof window<"u"?window.innerWidth<=768:!1);function N(){T.value=window.innerWidth<=768}Fe(()=>{window.addEventListener("resize",N),N()}),Ke(()=>{window.removeEventListener("resize",N)}),je(f,a=>{a||D(!1,"","")});function P(){r.value=p.username,h.value=p.password,f.value=p.remember}function F(){v.value="",u.value="",m.value=""}function V(){if(l.value==="login"){F(),l.value="register";return}P(),l.value="login"}async function E(){if(!r.value.trim()||!h.value.trim()){c.error("请完整填写登录信息");return}const a=await fetch(`${g.apiBase}/api/auth/login`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({username:r.value.trim(),password:h.value})}),o=await a.json().catch(()=>({}));if(!a.ok)throw new Error(o.error||"登录失败");D(f.value,r.value.trim(),h.value),g.setSession(o.token,o.username,String(o.expire_at||""),Number(o.session_idle_ttl_seconds||3600)),o.default_pwd&&c.warning("当前仍是默认密码，建议登录后尽快修改"),await s.push("/")}async function e(){if(!v.value.trim()||!u.value.trim()||!m.value.trim()){c.error("请完整填写注册信息");return}if(u.value!==m.value){c.error("两次输入密码不一致");return}if(u.value.trim().length<8){c.error("密码长度至少 8 位");return}const a=await fetch(`${g.apiBase}/api/auth/register`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({username:v.value.trim(),password:u.value})}),o=await a.json().catch(()=>({}));if(!a.ok)throw new Error(o.error||"注册失败");c.success("注册成功，请使用新账号登录"),l.value="login",v.value="",u.value="",m.value=""}async function k(){S.value=!0;try{l.value==="login"?await E():await e()}catch(a){c.error((a==null?void 0:a.message)||(l.value==="login"?"登录失败":"注册失败"))}finally{S.value=!1}}return(a,o)=>(M(),H("div",{class:"login-wrap",onKeyup:Oe(k,["enter"])},[z(He,{name:"auth-panel",mode:"out-in",appear:""},{default:C(()=>[(M(),H("div",{key:l.value,class:"auth-panel-wrap"},[z(b(We),{class:"login-card",title:l.value==="login"?"登录":"注册"},{default:C(()=>[Q("div",to,O(l.value==="login"?"欢迎使用运维集成管理后台系统":"创建管理员账号"),1),z(b(Je),{class:X(["login-form",l.value==="register"?"is-register":"is-login"]),"label-placement":T.value?"top":"left","label-align":T.value?"left":"right","label-width":T.value?void 0:96},{default:C(()=>[l.value==="login"?(M(),$(b(L),{key:0,label:"管理员账号"},{default:C(()=>[z(b(U),{value:r.value,"onUpdate:value":o[0]||(o[0]=t=>r.value=t),placeholder:"请输入账号"},null,8,["value"])]),_:1})):(M(),$(b(L),{key:1,label:"管理员账号"},{default:C(()=>[z(b(U),{value:v.value,"onUpdate:value":o[1]||(o[1]=t=>v.value=t),placeholder:"请输入账号"},null,8,["value"])]),_:1})),l.value==="login"?(M(),$(b(L),{key:2,label:"管理员密码"},{default:C(()=>[z(b(U),{value:h.value,"onUpdate:value":o[2]||(o[2]=t=>h.value=t),type:"password",placeholder:"请输入密码","show-password-on":"click"},null,8,["value"])]),_:1})):(M(),$(b(L),{key:3,label:"管理员密码"},{default:C(()=>[z(b(U),{value:u.value,"onUpdate:value":o[3]||(o[3]=t=>u.value=t),type:"password",placeholder:"请输入密码","show-password-on":"click"},null,8,["value"])]),_:1})),l.value==="register"?(M(),$(b(L),{key:4,label:"确认密码"},{default:C(()=>[z(b(U),{value:m.value,"onUpdate:value":o[4]||(o[4]=t=>m.value=t),type:"password",placeholder:"请再次输入密码","show-password-on":"click"},null,8,["value"])]),_:1})):Ge("",!0),Q("div",{class:X(["login-options",l.value==="register"?"is-register-options":"is-login-options"])},[l.value==="login"?(M(),$(b(no),{key:0,checked:f.value,"onUpdate:checked":o[5]||(o[5]=t=>f.value=t)},{default:C(()=>[...o[6]||(o[6]=[A("记住我",-1)])]),_:1},8,["checked"])):(M(),H("span",io)),z(b(Z),{text:"",type:"primary",class:"switch-link",onClick:V},{default:C(()=>[A(O(l.value==="login"?"还没有账号？立即注册":"已有账号？立即登录"),1)]),_:1})],2),z(b(Z),{class:"login-submit",block:"",type:"primary",loading:S.value,onClick:k},{default:C(()=>[A(O(l.value==="login"?"登录":"注册"),1)]),_:1},8,["loading"])]),_:1},8,["class","label-placement","label-align","label-width"])]),_:1},8,["title"])]))]),_:1})],32))}}),bo=qe(so,[["__scopeId","data-v-a9e40106"]]);export{bo as default};
