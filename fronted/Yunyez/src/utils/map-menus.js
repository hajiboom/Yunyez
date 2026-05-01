import router from '@/router';
import { lo } from 'element-plus/es/locales.mjs';


/**
 * 加载本地路由
 * @returns 所有本地路由数组
 */
function loadLocalRoutes(){
    const localRoutes=[];
const files = import.meta.glob('@/router/main/**/*.js', { eager: true }); 
   for(const key in files) {
        localRoutes.push(files[key].default);

    }
    return localRoutes;
}

export let firstMenu = null;

export const mapMenusToRoutes = (userMenus) => {
  const localRoutes = loadLocalRoutes();   // 加载本地定义的所有路由（如 /main/product/list）
  const routes = [];                        // 最终要添加的动态路由数组

  // 递归遍历菜单树，返回当前菜单项下第一个有效子路由的路径（供父级重定向使用）
  function traverse(menuList) {
    let firstChildPath = null;              // 记录本层第一个匹配到的子路由路径

    for (const menu of menuList) {
      if (menu.children && menu.children.length) {
        // 1. 非叶子节点：先递归处理子菜单
        const childFirstPath = traverse(menu.children);
        
        // 2. 如果子菜单中有有效路由，且当前父节点尚未添加重定向，则添加
        if (childFirstPath && !routes.find(route => route.path === menu.path)) {
          routes.push({
            path: menu.path,
            redirect: childFirstPath
          });
        }
        
        // 3. 记录本层第一个有效路径（用于上层父级）
        if (childFirstPath && firstChildPath === null) {
          firstChildPath = childFirstPath;
        }
      } else {
        // 叶子节点：匹配本地路由
        const matchedRoute = localRoutes.find(route => route.path === menu.url);
        if (matchedRoute) {
          // 添加该子路由
          routes.push(matchedRoute);
          
          // 记录全局第一个叶子菜单（用于默认激活）
          if (!firstMenu) {
            firstMenu = menu;
          }
          
          // 返回该叶子路径给父级
          if (firstChildPath === null) {
            firstChildPath = menu.path;
          }
        }
      }
    }
    
    return firstChildPath;   // 将本层第一个有效路径返回给上层调用
  }

  traverse(userMenus);
  return routes;
};

export function mapPathToMenu(path, userMenus) {
  for (const menu of userMenus) {
    // 当前菜单项匹配路径，直接返回
    if (menu.url === path) {
      return menu;
    }
    // 存在子菜单，递归查找
    if (menu.children && menu.children.length) {
      const result = mapPathToMenu(path, menu.children);
      if (result) {
        return result;
      }
    }
  }
  return null; // 未找到匹配项
}

/**
 * 递归遍历菜单列表，将所有子菜单的id添加到数组中
 * @param {*} menuList 菜单列表
 * @returns 所有子菜单的id数组
 */

export const mapMenuListToIds = (menuList) => {
  
    const ids = [];
  //递归遍历
  function traverse(menu) {
   for(const item of menu) {
    if(item.children) {
      traverse(item.children);
    }else{
        ids.push(item.id);
    }
   }
  }
  traverse(menuList);
  return ids;
}



/**
 * 按钮权限映射
 * @param {*} buttonList 按钮列表
 * @returns 按钮权限映射对象
 */
export const  mapMenuToPermissions = (buttonList) => {
    const permissions = {};
    //递归遍历按钮列表，将所有子按钮的id添加到对象中
    function  recurseGetPermission(button) {
       for(const item of button) {
        if(item.type===3) {// type === 3 表示按钮权限,选出类型为“按钮权限”的节点
            permissions.push(item.id);
        }else{
            recurseGetPermission(item.children?? []);
        }
       }
    }
    recurseGetPermission(buttonList);
    return permissions;
}



