"use strict";
'use client';

var _interopRequireDefault = require("@babel/runtime/helpers/interopRequireDefault");
Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = StyledEngineProvider;
var React = _interopRequireWildcard(require("react"));
var _propTypes = _interopRequireDefault(require("prop-types"));
var _react2 = require("@emotion/react");
var _cache = _interopRequireDefault(require("@emotion/cache"));
var _jsxRuntime = require("react/jsx-runtime");
function _getRequireWildcardCache(e) { if ("function" != typeof WeakMap) return null; var r = new WeakMap(), t = new WeakMap(); return (_getRequireWildcardCache = function (e) { return e ? t : r; })(e); }
function _interopRequireWildcard(e, r) { if (!r && e && e.__esModule) return e; if (null === e || "object" != typeof e && "function" != typeof e) return { default: e }; var t = _getRequireWildcardCache(r); if (t && t.has(e)) return t.get(e); var n = { __proto__: null }, a = Object.defineProperty && Object.getOwnPropertyDescriptor; for (var u in e) if ("default" !== u && Object.prototype.hasOwnProperty.call(e, u)) { var i = a ? Object.getOwnPropertyDescriptor(e, u) : null; i && (i.get || i.set) ? Object.defineProperty(n, u, i) : n[u] = e[u]; } return n.default = e, t && t.set(e, n), n; }
// prepend: true moves MUI styles to the top of the <head> so they're loaded first.
// It allows developers to easily override MUI styles with other styling solutions, like CSS modules.
function getCache(injectFirst, enableCssLayer) {
  const emotionCache = (0, _cache.default)({
    key: 'css',
    prepend: injectFirst
  });
  if (enableCssLayer) {
    const prevInsert = emotionCache.insert;
    emotionCache.insert = (...args) => {
      if (!args[1].styles.match(/^@layer\s+[^{]*$/)) {
        // avoid nested @layer
        args[1].styles = `@layer mui {${args[1].styles}}`;
      }
      return prevInsert(...args);
    };
  }
  return emotionCache;
}
const cacheMap = new Map();
function StyledEngineProvider(props) {
  const {
    injectFirst,
    enableCssLayer,
    children
  } = props;
  const cache = React.useMemo(() => {
    const cacheKey = `${injectFirst}-${enableCssLayer}`;
    if (typeof document === 'object' && cacheMap.has(cacheKey)) {
      return cacheMap.get(cacheKey);
    }
    const fresh = getCache(injectFirst, enableCssLayer);
    cacheMap.set(cacheKey, fresh);
    return fresh;
  }, [injectFirst, enableCssLayer]);
  if (injectFirst || enableCssLayer) {
    return /*#__PURE__*/(0, _jsxRuntime.jsx)(_react2.CacheProvider, {
      value: cache,
      children: children
    });
  }
  return children;
}
process.env.NODE_ENV !== "production" ? StyledEngineProvider.propTypes = {
  /**
   * Your component tree.
   */
  children: _propTypes.default.node,
  /**
   * If true, MUI styles are wrapped in CSS `@layer mui` rule.
   * It helps to override MUI styles when using CSS Modules, Tailwind CSS, plain CSS, or any other styling solution.
   */
  enableCssLayer: _propTypes.default.bool,
  /**
   * By default, the styles are injected last in the <head> element of the page.
   * As a result, they gain more specificity than any other style sheet.
   * If you want to override MUI's styles, set this prop.
   */
  injectFirst: _propTypes.default.bool
} : void 0;