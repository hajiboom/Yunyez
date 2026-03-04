import dayjs from 'dayjs'

/**
 * 全局通用时间格式化函数
 * @param {string|Date|null} time - 原始时间（ISO字符串/Date对象/空值）
 * @param {string} format - 格式化模板，默认 YYYY-MM-DD HH:mm:ss
 * @param {string} fallback - 时间无效时的兜底显示，默认 '未注册'
 * @returns {string} 格式化后的时间
 */
export const formatDateTime = (time, format = 'YYYY-MM-DD HH:mm:ss', fallback = '未注册') => {
  // 空值/无效值直接返回兜底
  if (!time || time === '') return fallback;
  
  // 解析时间并校验有效性
  const dayjsTime = dayjs(time);
  return dayjsTime.isValid() ? dayjsTime.format(format) : fallback;
};

/**
 * 简化版：仅格式化年月日
 * @param {string|Date|null} time - 原始时间
 * @returns {string} 格式化后的日期（YYYY-MM-DD）
 */
export const formatDate = (time) => {
  return formatDateTime(time, 'YYYY-MM-DD', '无日期');
};

