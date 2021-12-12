import { Params } from 'react-router'
import { HtmlMetaDescriptor } from 'remix'
import type { RootData } from './root'

export type MetaFunction<
  AppData = any,
  ParamsKey extends string = string,
  RouteData extends Record<string, any> = Record<string, any>,
> = (args: {
  data: AppData
  parentsData: RouteData
  params: Params<ParamsKey>
  location: Location
}) => HtmlMetaDescriptor

export type LoaderFunction<
  AppData = any,
  ParamsKey extends string = string,
  AppLoadContext = any,
> = (args: {
  request: Request
  context: AppLoadContext
  params: Params<ParamsKey>
}) => Promise<Response> | Response | Promise<AppData> | AppData

export type BlogMetaFunction<
  AppData = any,
  RouteData extends Record<string, any> = Record<string, any>,
  ParamsKey extends string = string,
> = MetaFunction<AppData, ParamsKey, RouteData & { root: RootData }>
