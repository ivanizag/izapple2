'use client';

import _typeof from "@babel/runtime/helpers/esm/typeof";
import * as React from 'react';
import PropTypes from 'prop-types';
import { CacheProvider } from '@emotion/react';
import createCache from '@emotion/cache';

// prepend: true moves MUI styles to the top of the <head> so they're loaded first.
// It allows developers to easily override MUI styles with other styling solutions, like CSS modules.
import { jsx as _jsx } from "react/jsx-runtime";
function getCache(injectFirst, enableCssLayer) {
  var emotionCache = createCache({
    key: 'css',
    prepend: injectFirst
  });
  if (enableCssLayer) {
    var prevInsert = emotionCache.insert;
    emotionCache.insert = function () {
      for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
        args[_key] = arguments[_key];
      }
      if (!args[1].styles.match(/^@layer\s+[^{]*$/)) {
        // avoid nested @layer
        args[1].styles = "@layer mui {".concat(args[1].styles, "}");
      }
      return prevInsert.apply(void 0, args);
    };
  }
  return emotionCache;
}
var cacheMap = new Map();
export default function StyledEngineProvider(props) {
  var injectFirst = props.injectFirst,
    enableCssLayer = props.enableCssLayer,
    children = props.children;
  var cache = React.useMemo(function () {
    var cacheKey = "".concat(injectFirst, "-").concat(enableCssLayer);
    if ((typeof document === "undefined" ? "undefined" : _typeof(document)) === 'object' && cacheMap.has(cacheKey)) {
      return cacheMap.get(cacheKey);
    }
    var fresh = getCache(injectFirst, enableCssLayer);
    cacheMap.set(cacheKey, fresh);
    return fresh;
  }, [injectFirst, enableCssLayer]);
  if (injectFirst || enableCssLayer) {
    return /*#__PURE__*/_jsx(CacheProvider, {
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
  children: PropTypes.node,
  /**
   * If true, MUI styles are wrapped in CSS `@layer mui` rule.
   * It helps to override MUI styles when using CSS Modules, Tailwind CSS, plain CSS, or any other styling solution.
   */
  enableCssLayer: PropTypes.bool,
  /**
   * By default, the styles are injected last in the <head> element of the page.
   * As a result, they gain more specificity than any other style sheet.
   * If you want to override MUI's styles, set this prop.
   */
  injectFirst: PropTypes.bool
} : void 0;